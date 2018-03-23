package fission

import (
	"fmt"
	"net/http"

	"k8s.io/apimachinery/pkg/util/validation"
)

type Resource interface {
	Validate() []string
}

func (archive Archive) Validate() (errs []string) {
	switch archive.Type {
	case ArchiveTypeLiteral, ArchiveTypeUrl: // no op
	default:
		return []string{fmt.Sprintf("%v is not a valid archive type", archive.Type)}
	}
	switch archive.Checksum.Type {
	case ChecksumTypeSHA256: // no op
	default:
		errs = append(errs, fmt.Sprintf("%v is not a valid checksum type", archive.Checksum.Type))
	}
	return errs
}

func (ref EnvironmentReference) Validate() (errs []string) {
	for _, s := range []string{ref.Name, ref.Namespace} {
		errs = append(errs, validation.IsDNS1123Label(s)...)
	}
	return errs
}

func (ref SecretReference) Validate() (errs []string) {
	for _, s := range []string{ref.Name, ref.Namespace} {
		errs = append(errs, validation.IsDNS1123Label(s)...)
	}
	return errs
}

func (ref ConfigMapReference) Validate() (errs []string) {
	for _, s := range []string{ref.Name, ref.Namespace} {
		errs = append(errs, validation.IsDNS1123Label(s)...)
	}
	return errs
}

func (spec PackageSpec) Validate() (errs []string) {
	for _, r := range []Resource{spec.Environment, spec.Source, spec.Deployment} {
		errs = append(errs, r.Validate()...)
	}
	return errs
}

func (ref PackageStatus) Validate() (errs []string) {
	switch ref.BuildStatus {
	case BuildStatusPending, BuildStatusRunning, BuildStatusSucceeded, BuildStatusFailed, BuildStatusNone: // no op
	default:
		errs = append(errs, fmt.Sprintf("%v is not a valid build status", ref.BuildStatus))
	}
	return errs
}

func (ref PackageRef) Validate() (errs []string) {
	for _, s := range []string{ref.Name, ref.Namespace} {
		errs = append(errs, validation.IsDNS1123Label(s)...)
	}
	return errs
}

func (ref FunctionPackageRef) Validate() (errs []string) {
	errs = append(errs, ref.PackageRef.Validate()...)
	errs = append(errs, validation.IsDNS1123Label(ref.FunctionName)...)
	return errs
}

func (spec FunctionSpec) Validate() (errs []string) {
	for _, r := range []Resource{spec.Environment, spec.Package} {
		errs = append(errs, r.Validate()...)
	}
	for _, s := range spec.Secrets {
		errs = append(errs, s.Validate()...)
	}
	for _, c := range spec.ConfigMaps {
		errs = append(errs, c.Validate()...)
	}
	errs = append(errs, spec.InvokeStrategy.Validate()...)
	return errs
}

func (is InvokeStrategy) Validate() (errs []string) {
	switch is.StrategyType {
	case StrategyTypeExecution: // no op
	default:
		errs = append(errs, fmt.Sprintf("%v is not a valid strategy type", is.StrategyType))
	}
	errs = append(errs, is.ExecutionStrategy.Validate()...)
	return errs
}

func (es ExecutionStrategy) Validate() (errs []string) {
	switch es.ExecutorType {
	case ExecutorTypeNewdeploy, ExecutorTypePoolmgr: // no op
	default:
		errs = append(errs, fmt.Sprintf("%v is not a valid executor type", es.ExecutorType))
	}
	if es.MinScale < 0 {
		errs = append(errs, "Minimum scale must be greater or equal to 0")
	}
	if es.MaxScale < es.MinScale {
		errs = append(errs, "Maximum scale must be greater or equal to minimum scale")
	}
	if es.TargetCPUPercent <= 0 || es.TargetCPUPercent > 100 {
		errs = append(errs, "TargetCPU must be a value between 1 - 100")
	}
	return errs
}

func (ref FunctionReference) Validate() (errs []string) {
	switch ref.Type {
	case FunctionReferenceTypeFunctionName: // no op
	default:
		errs = append(errs, fmt.Sprintf("%v is not a valid function reference type", ref.Type))
	}
	errs = append(errs, validation.IsDNS1123Label(ref.Name)...)
	return errs
}

func (runtime Runtime) Validate() (errs []string) {
	for _, port := range []int32{runtime.LoadEndpointPort, runtime.FunctionEndpointPort} {
		errs = append(errs, validation.IsValidPortNum(int(port))...)
	}
	return errs
}

func (builder Builder) Validate() (errs []string) {
	// do nothing for now
	return nil
}

func (spec EnvironmentSpec) Validate() (errs []string) {
	if spec.Version < 1 && spec.Version > 3 {
		errs = append(errs, "%v is not a valid environment version")
	}
	for _, r := range []Resource{spec.Runtime, spec.Builder} {
		errs = append(errs, r.Validate()...)
	}
	switch spec.AllowedFunctionsPerContainer {
	case AllowedFunctionsPerContainerSingle, AllowedFunctionsPerContainerInfinite: // no op
	default:
		errs = append(errs, fmt.Sprintf("%v is not a valid value", spec.AllowedFunctionsPerContainer))
	}
	if spec.Poolsize < 0 {
		errs = append(errs, "Poolsize must be greater or equal to 0")
	}
	return errs
}

func (spec HTTPTriggerSpec) Validate() (errs []string) {
	switch spec.Method {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace: // no op
	default:
		errs = append(errs, fmt.Sprintf("%v is not a valid HTTP method", spec.Method))
	}
	errs = append(errs, spec.FunctionReference.Validate()...)
	if len(spec.Host) > 0 {
		errs = append(errs, validation.IsDNS1123Subdomain(spec.Host)...)
	}
	return errs
}

func (spec KubernetesWatchTriggerSpec) Validate() (errs []string) {
	errs = append(errs, validation.IsDNS1123Label(spec.Namespace)...)

	switch spec.Type {
	case "POD", "SERVICE", "REPLICATIONCONTROLLER", "JOB":
	default:
		errs = append(errs, fmt.Sprintf("%v is not supported type", spec.Type))
	}

	for k, v := range spec.LabelSelector {
		errs = append(errs, validation.IsQualifiedName(k)...)
		errs = append(errs, validation.IsValidLabelValue(v)...)
	}

	errs = append(errs, spec.FunctionReference.Validate()...)
	return errs
}

func (spec MessageQueueTriggerSpec) Validate() (errs []string) {
	errs = append(errs, spec.FunctionReference.Validate()...)

	switch spec.MessageQueueType {
	case MessageQueueTypeNats, MessageQueueTypeASQ: // no op
	default:
		errs = append(errs, fmt.Sprintf("%v is not a valid message queue type", spec.MessageQueueType))
	}

	if !IsTopicValid(spec.MessageQueueType, spec.Topic) {
		errs = append(errs, fmt.Sprintf("%v is not a valid topic", spec.Topic))
	}

	if len(spec.ResponseTopic) > 0 && !IsTopicValid(spec.MessageQueueType, spec.ResponseTopic) {
		errs = append(errs, fmt.Sprintf("%v is not a valid topic", spec.ResponseTopic))
	}

	return errs
}

func (spec TimeTriggerSpec) Validate() (errs []string) {
	err := IsValidCronSpec(spec.Cron)
	if err != nil {
		errs = append(errs, err.Error())
	}
	errs = append(errs, spec.FunctionReference.Validate()...)
	return errs
}
