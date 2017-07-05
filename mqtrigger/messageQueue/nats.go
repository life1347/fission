/*
Copyright 2016 The Fission Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package messageQueue

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	ns "github.com/nats-io/go-nats-streaming"
	nsUtil "github.com/nats-io/nats-streaming-server/util"
	log "github.com/sirupsen/logrus"

	"github.com/fission/fission"
)

const (
	natsClusterID  = "fissionMQTrigger"
	natsProtocol   = "nats://"
	natsClientID   = "fission"
	natsQueueGroup = "fission-messageQueueNatsTrigger"
)

type (
	Nats struct {
		nsConn    ns.Conn
		routerUrl string
	}
)

func makeNatsMessageQueue(routerUrl string, mqCfg MessageQueueConfig) (MessageQueue, error) {
	conn, err := ns.Connect(natsClusterID, natsClientID, ns.NatsURL(mqCfg.Url))
	if err != nil {
		return nil, err
	}
	nats := Nats{
		nsConn:    conn,
		routerUrl: routerUrl,
	}
	return nats, nil
}

func (nats Nats) subscribe(trigger fission.MessageQueueTrigger) (messageQueueSubscription, error) {
	subj := trigger.Topic

	if !isTopicValidForNats(subj) {
		return nil, errors.New(fmt.Sprintf("Not a valid topic: %s", trigger.Topic))
	}

	opts := []ns.SubscriptionOption{
		// Create a durable subscription to nats, so that triggers could retrieve last unack message.
		// https://github.com/nats-io/go-nats-streaming#durable-subscriptions
		ns.DurableName(trigger.Uid),

		// Nats-streaming server is auto-ack mode by default. Since we want nats-streaming server to
		// resend a message if the trigger does not ack it, we need to enable the manual ack mode, so that
		// trigger could choose to ack message or simply drop it depend on the response of function pod.
		ns.SetManualAckMode(),
	}
	sub, err := nats.nsConn.Subscribe(subj, msgHandler(&nats, trigger), opts...)
	if err != nil {
		return nil, err
	}
	return sub, nil
}

func (nats Nats) unsubscribe(subscription messageQueueSubscription) error {
	return subscription.(ns.Subscription).Close()
}

func isTopicValidForNats(topic string) bool {
	// nats-streaming does not support wildcard channel.
	return nsUtil.IsSubjectValid(topic)
}

func msgHandler(nats *Nats, trigger fission.MessageQueueTrigger) func(*ns.Msg) {
	return func(msg *ns.Msg) {
		url := nats.routerUrl + "/" + strings.TrimPrefix(fission.UrlForFunction(&trigger.Function), "/")
		log.Printf("Making HTTP request to %v", url)

		headers := map[string]string{
			"X-Fission-MQTrigger-Topic":     trigger.Topic,
			"X-Fission-MQTrigger-RespTopic": trigger.ResponseTopic,
		}

		// Create request
		req, err := http.NewRequest("POST", url, bytes.NewReader(msg.Data))
		for k, v := range headers {
			req.Header.Add(k, v)
		}

		// Make the request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Warningf("Request failed: %v", url)
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Warningf("Request body error: %v", string(body))
			return
		}
		if resp.StatusCode != 200 {
			log.Printf("Request returned failure: %v", resp.StatusCode)
			return
		}
		// trigger acks message only if a request done successfully
		err = msg.Ack()
		if err != nil {
			log.Warningf("Failed to ack message: %v", err)
		}
		err = nats.nsConn.Publish(trigger.ResponseTopic, body)
		if err != nil {
			log.Warningf("Failed to publish message to topic %s: %v", trigger.ResponseTopic, err)
		}
	}
}
