package v2vvmware

import (
	"context",
	"errors",

	kubevirtv1alpha1 "kubevirt.io/v2v-vmware/pkg/apis/kubevirt/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const PhaseConnecting = "Connecting"
const PhaseConnectionSuccessful = "True"
const PhaseConnectionFailed = "Failed"

var log = logf.Log.WithName("controller_v2vvmware")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new V2VVmware Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileV2VVmware{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("v2vvmware-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource V2VVmware
	err = c.Watch(&source.Kind{Type: &kubevirtv1alpha1.V2VVmware{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
/*
	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner V2VVmware
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &kubevirtv1alpha1.V2VVmware{},
	})
	if err != nil {
		return err
	}
*/
	return nil
}

var _ reconcile.Reconciler = &ReconcileV2VVmware{}

// ReconcileV2VVmware reconciles a V2VVmware object
type ReconcileV2VVmware struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a V2VVmware object and makes changes based on the state read
// and what is in the V2VVmware.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileV2VVmware) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling V2VVmware")

	// Fetch the V2VVmware instance
	instance := &kubevirtv1alpha1.V2VVmware{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

    connectionSecret, err := getConnectionSecret(r, request, instance)
    if err != nil {
    	reqLogger.Error(err, "Failed to get Secret object for the VMWare connection")
		return reconcile.Result{}, err // request will be re-queued
	}

    if instance.Spec.Vms == nil { // if missing at all, then just connection check is requested
    	err = checkConnectionOnly(r, request, instance, connectionSecret)
    	if err != nil {
			reqLogger.Error(err, "Failed to check VMWare connection.")
		}
		return reconcile.Result{}, err // request will be re-queued if failed
	}

    // Considering recent high-level flow, the list of VMWare VMs is read at most once (means: do not refresh).
    // If refresh is ever needed, implement either here or re-create the V2VVmware object

	if len(instance.Spec.Vms) == 0 { // list of VMWare VMs is requested to be retrieved
		err = readVmsList(r, request, connectionSecret)
		if err != nil {
			reqLogger.Error(err, "Failed to read list of VMWare VMs.")
		}
		return reconcile.Result{}, err // request will be re-queued if failed
	}

    // secret is present, list of VMs is available, let's check for  details to be retrieved
    var lastError = nil
    for _, vm := range instance.Spec.Vms { // sequential read is probably good enough, just a single VM or a few of them are expected to be retrieved this way
    	if vm.DetailRequest {
			err = readVmDetail(r, request, connectionSecret, vm)
			if err != nil {
				reqLogger.Error(err, fmt.Sprintf("Failed to read detail of '%s' VMWare VM.", vm.name))
				lastError = err
			}
		}
	}

	return reconcile.Result{}, lastError
}

func getConnectionSecret(r *ReconcileV2VVmware, request reconcile.Request, instance *kubevirtv1alpha1.V2VVmware) (*corev1.Secret, error) {
    if instance.Spec.Connection == "" {
        return nil, errors.New("The Spec.Connection is required in a V2VVmware object. References a Secret by name.")
    }

    secret := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.Connection, Namespace: request.Namespace}, secret)
    return secret, err
}

func checkConnectionOnly(r *ReconcileV2VVmware, request reconcile.Request, instance *kubevirtv1alpha1.V2VVmware) (error) {
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

	instance.Spec.Vms = make([]VmwareVm, len(vmwareVms))
	for index, vmName := range vmwareVms {
		instance.Spec.Vms[index] = kubevirtv1alpha1.VmwareVm{
			Name: vmName,
			DetailRequest: false, // can be omitted, but just to be clear
			Detail: nil,
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
		log.Error(err, fmt.Sprintf("Failed to update V2VVmware object with detail of '%s' VM.", vmwareVms))
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
		log.Error(err, fmt.Sprintf("Failed to update V2VVmware status. Intended to write phase: '%s', message: %s", phase, msg))
	}
}

