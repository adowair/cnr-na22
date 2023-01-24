We are going to develop a CRD based API to manage docker images in a repository.
The primary resource in this API will be a `Build` object. Builds represent
built images that have been pushed to some repository. The existence of a
`Build` object represents the desire for there to exist some specific image in
some specific repository. The Build CR controller is responsible for reconciling
the state of these Build object. This includes checking whether the registry
already has the image, and if not, building and pushing it.

1. Initialize a go project:
   ```shell
   $ go mod init adowair.github.io/cicd
   ```
2. Initialize kubebuilder project:
   ```shell
   $ kubebuilder init
   Writing kustomize manifests for you to edit...
   Writing scaffold for you to edit...
   ```
3. Create a new API with resource and controller:
   ```shell
   $ kubebuilder create api --group cicd --version v1alpha1 --kind Build
   Create Resource [y/n]
   > y
   Create Controller [y/n]
   > y
   ```
4. Add wanted fields to `Build`'s status and spec in
   [build_types.go](api/v1alpha1/build_types.go). After modifying this file,
   remember to run `make` to regenerate controller code. These status and spec
   structs will initially look like this:
   ```go
   // BuildSpec defines the desired state of Build
   type BuildSpec struct {
     // Foo is an example field of Build. Edit build_types.go to remove/update
     Foo string `json:"foo,omitempty"`
   }

   // BuildStatus defines the observed state of Build
   type BuildStatus struct {
     // INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
   	 // Important: Run "make" to regenerate code after modifying this file
   }
   ```
5. Add reconcile loop logic to [build_controller.go](controllers/build_controller.go).
   The function `BuildReconciler.Reconcile()` will run whenever Kubernetes client
   makes a request against this API.
   ```go
   func (r *BuildReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	 _ = log.FromContext(ctx)

	 // TODO(user): your logic here

	 return ctrl.Result{}, nil
   }
   ```
   