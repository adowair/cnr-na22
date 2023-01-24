/*
Copyright 2023.

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

package controllers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cicdv1alpha1 "adowair.github.io/cicd/api/v1alpha1"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// BuildReconciler reconciles a Build object
type BuildReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cicd.adowair.github.io,resources=builds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cicd.adowair.github.io,resources=builds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cicd.adowair.github.io,resources=builds/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Build object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *BuildReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Get the Build object
	var build cicdv1alpha1.Build
	if err := r.Get(ctx, req.NamespacedName, &build); err != nil {
		log.Error(err, "unable to fetch Build")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Create a docker client to the specified host
	client, err := docker.NewClientWithOpts(docker.WithHost(build.Spec.Host))
	if err != nil {
		log.Error(err, "unable to create docker client")
		return ctrl.Result{}, err
	}

	// Check that the registry doesn't already have this build
	filter := filters.NewArgs(filters.Arg("reference", fmt.Sprintf("%s:%s", build.Spec.Image, build.Spec.Tag)))
	images, err := client.ImageList(ctx, types.ImageListOptions{Filters: filter})
	if err != nil {
		log.Error(err, "unable to list images in registry")
		return ctrl.Result{}, err
	}

	if len(images) > 0 {
		log.Info("an image with the specified name an tag already exists", "images", images)
		return ctrl.Result{}, errors.New("an image with the specified tag already exists")
	}

	repoDir, err := os.MkdirTemp("", "git-repo-")
	if err != nil {
		return ctrl.Result{}, err
	}
	defer os.RemoveAll(repoDir)

	repo, err := git.PlainClone(repoDir, true, &git.CloneOptions{
		URL: build.Spec.Host,
	})
	if err != nil {
		log.Error(err, "unable to clone repository")
		return ctrl.Result{}, errors.New("unable to clone repository")
	}

	w, err := repo.Worktree()
	if err != nil {
		return ctrl.Result{}, err
	}

	w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(build.Spec.Commit),
	})

	dockerfilePath := filepath.Join(repoDir, build.Spec.Dockerfile)

	_, err = client.ImageBuild(ctx, nil, types.ImageBuildOptions{
		Dockerfile: dockerfilePath,
		Tags:       []string{build.Spec.Tag},
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	// TODO: push image to registry

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BuildReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cicdv1alpha1.Build{}).
		Complete(r)
}
