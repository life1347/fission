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

package crd

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation"

	"github.com/fission/fission"
)

//
// To add a Fission CRD type:
//   1. Create a "spec" type, for everything in the type except metadata
//   2. Create the type with metadata + the spec
//   3. Create a list type (for example see FunctionList and Function, below)
//   4. Add methods at the bottom of this file for satisfying Object and List interfaces
//   5. Add the type to configureClient in client.go
//   6. Add the type to EnsureFissionCRDs in crd.go
//   7. Add tests to crd_test.go
//   8. Add a CRUD Interface type (analogous to FunctionInterface in function.go)
//   9. Add a getter method for your interface type to FissionClient in client.go
//

type (
	// Packages. Think of these as function-level images.
	Package struct {
		metav1.TypeMeta `json:",inline"`
		Metadata        metav1.ObjectMeta   `json:"metadata"`
		Spec            fission.PackageSpec `json:"spec"`

		Status fission.PackageStatus `json:"status"`
	}
	PackageList struct {
		metav1.TypeMeta `json:",inline"`
		Metadata        metav1.ListMeta `json:"metadata"`

		Items []Package `json:"items"`
	}

	// Functions.
	Function struct {
		metav1.TypeMeta `json:",inline"`
		Metadata        metav1.ObjectMeta    `json:"metadata"`
		Spec            fission.FunctionSpec `json:"spec"`
	}
	FunctionList struct {
		metav1.TypeMeta `json:",inline"`
		Metadata        metav1.ListMeta `json:"metadata"`

		Items []Function `json:"items"`
	}

	// Environments.
	Environment struct {
		metav1.TypeMeta `json:",inline"`
		Metadata        metav1.ObjectMeta       `json:"metadata"`
		Spec            fission.EnvironmentSpec `json:"spec"`
	}
	EnvironmentList struct {
		metav1.TypeMeta `json:",inline"`
		Metadata        metav1.ListMeta `json:"metadata"`

		Items []Environment `json:"items"`
	}

	HTTPTrigger struct {
		metav1.TypeMeta `json:",inline"`
		Metadata        metav1.ObjectMeta       `json:"metadata"`
		Spec            fission.HTTPTriggerSpec `json:"spec"`
	}
	HTTPTriggerList struct {
		metav1.TypeMeta `json:",inline"`
		Metadata        metav1.ListMeta `json:"metadata"`

		Items []HTTPTrigger `json:"items"`
	}

	// Kubernetes Watches as triggers
	KubernetesWatchTrigger struct {
		metav1.TypeMeta `json:",inline"`
		Metadata        metav1.ObjectMeta                  `json:"metadata"`
		Spec            fission.KubernetesWatchTriggerSpec `json:"spec"`
	}
	KubernetesWatchTriggerList struct {
		metav1.TypeMeta `json:",inline"`
		Metadata        metav1.ListMeta `json:"metadata"`

		Items []KubernetesWatchTrigger `json:"items"`
	}

	// Time triggers
	TimeTrigger struct {
		metav1.TypeMeta `json:",inline"`
		Metadata        metav1.ObjectMeta       `json:"metadata"`
		Spec            fission.TimeTriggerSpec `json:"spec"`
	}
	TimeTriggerList struct {
		metav1.TypeMeta `json:",inline"`
		Metadata        metav1.ListMeta `json:"metadata"`

		Items []TimeTrigger `json:"items"`
	}

	// Message Queue triggers
	MessageQueueTrigger struct {
		metav1.TypeMeta `json:",inline"`
		Metadata        metav1.ObjectMeta               `json:"metadata"`
		Spec            fission.MessageQueueTriggerSpec `json:"spec"`
	}
	MessageQueueTriggerList struct {
		metav1.TypeMeta `json:",inline"`
		Metadata        metav1.ListMeta `json:"metadata"`

		Items []MessageQueueTrigger `json:"items"`
	}
)

// Each CRD type needs:
//   GetObjectKind (to satisfy the Object interface)
//
// In addition, each singular CRD type needs:
//   GetObjectMeta (to satisfy the ObjectMetaAccessor interface)
//
// And each list CRD type needs:
//   GetListMeta (to satisfy the ListMetaAccessor interface)

func (f *Function) GetObjectKind() schema.ObjectKind {
	return &f.TypeMeta
}
func (e *Environment) GetObjectKind() schema.ObjectKind {
	return &e.TypeMeta
}
func (ht *HTTPTrigger) GetObjectKind() schema.ObjectKind {
	return &ht.TypeMeta
}
func (w *KubernetesWatchTrigger) GetObjectKind() schema.ObjectKind {
	return &w.TypeMeta
}
func (t *TimeTrigger) GetObjectKind() schema.ObjectKind {
	return &t.TypeMeta
}
func (m *MessageQueueTrigger) GetObjectKind() schema.ObjectKind {
	return &m.TypeMeta
}
func (p *Package) GetObjectKind() schema.ObjectKind {
	return &p.TypeMeta
}

func (f *Function) GetObjectMeta() metav1.Object {
	return &f.Metadata
}
func (e *Environment) GetObjectMeta() metav1.Object {
	return &e.Metadata
}
func (ht *HTTPTrigger) GetObjectMeta() metav1.Object {
	return &ht.Metadata
}
func (w *KubernetesWatchTrigger) GetObjectMeta() metav1.Object {
	return &w.Metadata
}
func (t *TimeTrigger) GetObjectMeta() metav1.Object {
	return &t.Metadata
}
func (m *MessageQueueTrigger) GetObjectMeta() metav1.Object {
	return &m.Metadata
}
func (p *Package) GetObjectMeta() metav1.Object {
	return &p.Metadata
}

func (fl *FunctionList) GetObjectKind() schema.ObjectKind {
	return &fl.TypeMeta
}
func (el *EnvironmentList) GetObjectKind() schema.ObjectKind {
	return &el.TypeMeta
}
func (hl *HTTPTriggerList) GetObjectKind() schema.ObjectKind {
	return &hl.TypeMeta
}
func (wl *KubernetesWatchTriggerList) GetObjectKind() schema.ObjectKind {
	return &wl.TypeMeta
}
func (wl *TimeTriggerList) GetObjectKind() schema.ObjectKind {
	return &wl.TypeMeta
}
func (ml *MessageQueueTriggerList) GetObjectKind() schema.ObjectKind {
	return &ml.TypeMeta
}
func (pl *PackageList) GetObjectKind() schema.ObjectKind {
	return &pl.TypeMeta
}

func (fl *FunctionList) GetListMeta() metav1.List {
	return &fl.Metadata
}
func (el *EnvironmentList) GetListMeta() metav1.List {
	return &el.Metadata
}
func (hl *HTTPTriggerList) GetListMeta() metav1.List {
	return &hl.Metadata
}
func (wl *KubernetesWatchTriggerList) GetListMeta() metav1.List {
	return &wl.Metadata
}
func (wl *TimeTriggerList) GetListMeta() metav1.List {
	return &wl.Metadata
}
func (ml *MessageQueueTriggerList) GetListMeta() metav1.List {
	return &ml.Metadata
}
func (pl *PackageList) GetListMeta() metav1.List {
	return &pl.Metadata
}

func validateMetadata(m metav1.ObjectMeta) (errs []string) {
	for _, s := range []string{m.Name, m.Namespace} {
		errs = append(errs, validation.IsDNS1123Label(s)...)
	}
	return errs
}

func (p *Package) Validate() (errs []string) {
	errs = append(errs, validateMetadata(p.Metadata)...)
	errs = append(errs, p.Spec.Validate()...)
	errs = append(errs, p.Status.Validate()...)
	return errs
}

func (pl *PackageList) Validate() (errs []string) {
	// not validate ListMeta
	for _, p := range pl.Items {
		errs = append(errs, p.Validate()...)
	}
	return errs
}

func (f *Function) Validate() (errs []string) {
	errs = append(errs, validateMetadata(f.Metadata)...)
	errs = append(errs, f.Spec.Validate()...)
	return errs
}

func (fl *FunctionList) Validate() (errs []string) {
	for _, f := range fl.Items {
		errs = append(errs, f.Validate()...)
	}
	return errs
}

func (e *Environment) Validate() (errs []string) {
	errs = append(errs, validateMetadata(e.Metadata)...)
	errs = append(errs, e.Spec.Validate()...)
	return errs
}

func (el *EnvironmentList) Validate() (errs []string) {
	for _, e := range el.Items {
		errs = append(errs, e.Validate()...)
	}
	return errs
}

func (h *HTTPTrigger) Validate() (errs []string) {
	errs = append(errs, validateMetadata(h.Metadata)...)
	errs = append(errs, h.Spec.Validate()...)
	return errs
}

func (hl *HTTPTriggerList) Validate() (errs []string) {
	for _, h := range hl.Items {
		errs = append(errs, h.Validate()...)
	}
	return errs
}

func (k *KubernetesWatchTrigger) Validate() (errs []string) {
	errs = append(errs, validateMetadata(k.Metadata)...)
	errs = append(errs, k.Spec.Validate()...)
	return errs
}

func (kl *KubernetesWatchTriggerList) Validate() (errs []string) {
	for _, k := range kl.Items {
		errs = append(errs, k.Validate()...)
	}
	return errs
}

func (t *TimeTrigger) Validate() (errs []string) {
	errs = append(errs, validateMetadata(t.Metadata)...)
	errs = append(errs, t.Spec.Validate()...)
	return errs
}

func (tl *TimeTriggerList) Validate() (errs []string) {
	for _, t := range tl.Items {
		errs = append(errs, t.Validate()...)
	}
	return errs
}

func (m *MessageQueueTrigger) Validate() (errs []string) {
	errs = append(errs, validateMetadata(m.Metadata)...)
	errs = append(errs, m.Spec.Validate()...)
	return errs
}

func (ml *MessageQueueTriggerList) Validate() (errs []string) {
	for _, m := range ml.Items {
		errs = append(errs, m.Validate()...)
	}
	return errs
}
