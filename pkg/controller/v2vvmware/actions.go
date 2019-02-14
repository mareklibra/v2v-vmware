package v2vvmware

import (
	"context"
	"errors"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	kubevirtv1alpha1 "kubevirt.io/v2v-vmware/pkg/apis/kubevirt/v1alpha1"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const MaxRetryCount = 5

func getConnectionSecret(r *ReconcileV2VVmware, request reconcile.Request, instance *kubevirtv1alpha1.V2VVmware) (*corev1.Secret, error) {
	if instance.Spec.Connection == "" {
		return nil, errors.New("the Spec.Connection is required in a V2VVmware object. References a Secret by name")
	}

	secret := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.Connection, Namespace: request.Namespace}, secret)
	return secret, err
}

func sleepBeforeRetry() {
	log.Info("Falling asleep before retry ...")
	time.Sleep(5 * time.Second)
}

/*
func checkConnectionOnly(r *ReconcileV2VVmware, request reconcile.Request, connectionSecret *corev1.Secret) (error) {
	log.Info("checkConnectionOnly()")
	updateStatusPhase(r, request, PhaseConnecting)

	time.Sleep(5 * time.Second)

	// TODO: verify connection to VMWare
	if true {
		updateStatusPhase(r, request, PhaseConnectionSuccessful)
	} else {
		updateStatusPhase(r, request, PhaseConnectionFailed)
	}
	return nil // TODO
}
*/
// read whole list at once
func readVmsList(r *ReconcileV2VVmware, request reconcile.Request, connectionSecret *corev1.Secret) error {
	log.Info("readVmsList()")

	updateStatusPhase(r, request, PhaseConnecting)
	time.Sleep(5 * time.Second) // Mimic connection check
	updateStatusPhase(r, request, PhaseConnectionSuccessful)

	// TODO: read following list from VMWare
	time.Sleep(10 * time.Second) // Mimic data retrieval delay
	vmwareVms := []string{"fake_vm_1", "fake_vm_2", "fake_vm_3"}

	return updateVmsList(r, request, vmwareVms, MaxRetryCount)
}

func updateVmsList(r *ReconcileV2VVmware, request reconcile.Request, vmwareVms []string, retryCount int) error {
	instance := &kubevirtv1alpha1.V2VVmware{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		log.Error(err, fmt.Sprintf("Failed to get V2VVmware object to update list of VMs, intended to write: '%s'", vmwareVms))
		if retryCount > 0 {
			sleepBeforeRetry()
			return updateVmsList(r, request, vmwareVms, retryCount - 1)
		}
		return err
	}

	instance.Spec.Vms = make([]kubevirtv1alpha1.VmwareVm, len(vmwareVms))
	for index, vmName := range vmwareVms {
		instance.Spec.Vms[index] = kubevirtv1alpha1.VmwareVm{
			Name:          vmName,
			DetailRequest: false, // can be omitted, but just to be clear
		}
	}

	err = r.client.Update(context.TODO(), instance)
	if err != nil {
		log.Error(err, fmt.Sprintf("Failed to update V2VVmware object with list of VMWare VMs, intended to write: '%s'", vmwareVms))
		if retryCount > 0 {
			sleepBeforeRetry()
			return updateVmsList(r, request, vmwareVms, retryCount - 1)
		}
		return err
	}

	return nil
}

func readVmDetail(r *ReconcileV2VVmware, request reconcile.Request, connectionSecret *corev1.Secret, vmwareVm *kubevirtv1alpha1.VmwareVm) (error) {
	log.Info("readVmDetail()")

	// TODO: read details of a single VM from VMWare (use vmwareVm.Name)
	vmDetail := kubevirtv1alpha1.VmwareVmDetail{
		DummyAttr: "foo",
	}

	instance := &kubevirtv1alpha1.V2VVmware{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		log.Error(err, fmt.Sprintf("Failed to get V2VVmware object to update detail of '%s' VM.", vmwareVm.Name))
		return err
	}

	for index, vm := range instance.Spec.Vms {
		if  vm.Name == vmwareVm.Name {
			instance.Spec.Vms[index].DetailRequest = false // skip this next time
			instance.Spec.Vms[index].Detail = vmDetail
		}
	}

	err = r.client.Update(context.TODO(), instance)
	if err != nil {
		log.Error(err, fmt.Sprintf("Failed to update V2VVmware object with detail of '%s' VM.", vmwareVm.Name))
		return err
	}

	return nil
}

func updateStatusPhase(r *ReconcileV2VVmware, request reconcile.Request, phase string) {
	log.Info(fmt.Sprintf("updateStatusPhase(): %s", phase))
	updateStatusPhaseRetry(r, request, phase, MaxRetryCount)
}

func updateStatusPhaseRetry(r *ReconcileV2VVmware, request reconcile.Request, phase string, retryCount int) {
	// reload instance to workaround issues with parallel writes
	instance := &kubevirtv1alpha1.V2VVmware{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		log.Error(err, fmt.Sprintf("Failed to get V2VVmware object to update status info. Intended to write phase: '%s'", phase))
		if retryCount > 0 {
			sleepBeforeRetry()
			updateStatusPhaseRetry(r, request, phase, retryCount - 1)
		}
		return
	}

	instance.Status.Phase = phase
	err = r.client.Update(context.TODO(), instance)
	if err != nil {
		log.Error(err, fmt.Sprintf("Failed to update V2VVmware status. Intended to write phase: '%s'", phase))
		if retryCount > 0 {
			sleepBeforeRetry()
			updateStatusPhaseRetry(r, request, phase, retryCount - 1)
		}
	}
}
