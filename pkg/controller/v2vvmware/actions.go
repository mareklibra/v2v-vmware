package v2vvmware

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	kubevirtv1alpha1 "kubevirt.io/v2v-vmware/pkg/apis/kubevirt/v1alpha1"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func getConnectionSecret(r *ReconcileV2VVmware, request reconcile.Request, instance *kubevirtv1alpha1.V2VVmware) (*corev1.Secret, error) {
	if instance.Spec.Connection == "" {
		return nil, errors.New("The Spec.Connection is required in a V2VVmware object. References a Secret by name.")
	}

	secret := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.Connection, Namespace: request.Namespace}, secret)
	return secret, err
}

func checkConnectionOnly(r *ReconcileV2VVmware, request reconcile.Request, connectionSecret *corev1.Secret) (error) {
	updateStatusPhase(r, request, PhaseConnecting)

	// TODO: verify connection to VMWare
	if true {
		updateStatusPhase(r, request, PhaseConnectionSuccessful)
	} else {
		updateStatusPhase(r, request, PhaseConnectionFailed)
	}
	return nil // TODO
}

// read whole list at once
func readVmsList(r *ReconcileV2VVmware, request reconcile.Request, connectionSecret *corev1.Secret) (error) {
	// TODO: read the list from VMWare
	vmwareVms := []string{"fake_vm_1", "fake_vm_2", "fake_vm_3"}

	instance := &kubevirtv1alpha1.V2VVmware{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		log.Error(err, fmt.Sprintf("Failed to get V2VVmware object to update list of VMs, intended to write: '%s'", vmwareVms))
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
		return err
	}

	return nil
}

func readVmDetail(r *ReconcileV2VVmware, request reconcile.Request, connectionSecret *corev1.Secret, vmwareVm *kubevirtv1alpha1.VmwareVm) (error) {
	// TODO: read details of a single VM from VMWare (use vmwareVm.Name)
	vmDetail := kubevirtv1alpha1.VmwareVmDetail{
		// TODO: set fields
	}

	instance := &kubevirtv1alpha1.V2VVmware{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		log.Error(err, fmt.Sprintf("Failed to get V2VVmware object to update detail of '%s' VM.", vmwareVm.Name))
		return err
	}

	for _, vm := range instance.Spec.Vms {
		if vm.Name == vmwareVm.Name {
			vm.DetailRequest = false // skip this next time
			vm.Detail = vmDetail
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
	// reload instance to workaround issues with parallel writes
	instance := &kubevirtv1alpha1.V2VVmware{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		log.Error(err, fmt.Sprintf("Failed to get V2VVmware object to update status info. Intended to write phase: '%s'", phase))
		return
	}

	instance.Status.Phase = phase
	err = r.client.Update(context.TODO(), instance)
	if err != nil {
		log.Error(err, fmt.Sprintf("Failed to update V2VVmware status. Intended to write phase: '%s', message: %s", phase))
	}
}
