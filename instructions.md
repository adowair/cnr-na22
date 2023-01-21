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