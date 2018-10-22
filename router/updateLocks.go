package router

import (
	"sync"
	"time"

	"github.com/fission/fission/crd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type updateLockOperation int

const (
	GET updateLockOperation = iota
	DELETE
	EXPIRE
)

type (
	updateLocks struct {
		requestChan chan *fnUpdateRequest
		locks       map[string]*updateLock
	}

	updateLock struct {
		wg        *sync.WaitGroup
		timestamp time.Time
	}

	fnUpdateRequest struct {
		requestType  updateLockOperation
		responseChan chan *fnUpdateResponse
		key          string
	}

	fnUpdateResponse struct {
		lock         *updateLock
		ableToUpdate bool
	}
)

func (l *updateLock) isOld() bool {
	return time.Since(l.timestamp) > 30*time.Second
}

func (l *updateLock) Wait() {
	l.wg.Wait()
}

func MakeUpdateLocks() *updateLocks {
	locks := &updateLocks{
		requestChan: make(chan *fnUpdateRequest),
		locks:       make(map[string]*updateLock),
	}
	go locks.service()
	return locks
}

func (ul *updateLocks) service() {

	for {
		req := <-ul.requestChan

		switch req.requestType {
		case GET:
			lock, ok := ul.locks[req.key]
			if ok && !lock.isOld() {
				req.responseChan <- &fnUpdateResponse{
					lock: lock, ableToUpdate: false,
				}
				continue
			} else if ok && lock.isOld() {
				// in case that one goroutine occupy the update lock for long time
				lock.wg.Done()
			}

			lock = &updateLock{
				wg:        &sync.WaitGroup{},
				timestamp: time.Now(),
			}

			lock.wg.Add(1)

			ul.locks[req.key] = lock

			req.responseChan <- &fnUpdateResponse{
				lock: lock, ableToUpdate: true,
			}

		case DELETE:
			lock, ok := ul.locks[req.key]
			if ok {
				lock.wg.Done()
				delete(ul.locks, req.key)
			}

		case EXPIRE:
			for k, v := range ul.locks {
				if v.isOld() {
					delete(ul.locks, k)
				}
			}
		}
	}
}

func (locks *updateLocks) Get(fnMeta *metav1.ObjectMeta) (lock *updateLock, ableToUpdate bool) {
	ch := make(chan *fnUpdateResponse)
	locks.requestChan <- &fnUpdateRequest{
		requestType:  GET,
		responseChan: ch,
		key:          crd.CacheKey(fnMeta),
	}
	resp := <-ch
	return resp.lock, resp.ableToUpdate
}

func (locks *updateLocks) Delete(fnMeta *metav1.ObjectMeta) {
	locks.requestChan <- &fnUpdateRequest{
		requestType: DELETE,
		key:         crd.CacheKey(fnMeta),
	}
}

func (locks *updateLocks) expiryService() {
	for {
		time.Sleep(time.Minute)
		locks.requestChan <- &fnUpdateRequest{
			requestType: EXPIRE,
		}
	}
}
