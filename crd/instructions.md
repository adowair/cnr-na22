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