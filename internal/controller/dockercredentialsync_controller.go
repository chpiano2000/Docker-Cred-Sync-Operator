/*
Copyright 2025.

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

package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	daxvov1 "dax.vo/docker-credentail-sync/api/v1"
)

// DockerCredentialSyncReconciler reconciles a DockerCredentialSync object
type DockerCredentialSyncReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=dax.vo,resources=dockercredentialsyncs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=dax.vo,resources=dockercredentialsyncs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=dax.vo,resources=dockercredentialsyncs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *DockerCredentialSyncReconciler) Reconcile(
	ctx context.Context,
	req ctrl.Request,
) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling Docker Credential Sync")
	now := metav1.Now()

	var dockerCredCR daxvov1.DockerCredentialSync
	if err := r.Get(ctx, req.NamespacedName, &dockerCredCR); err != nil {
		log.Error(err, "unable to fetch DockerCredentialSync")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Get source secret
	var sourceSecret corev1.Secret
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: dockerCredCR.Spec.SourceNamespace,
		Name:      dockerCredCR.Spec.SourceSecretName,
	}, &sourceSecret); err != nil {
		log.Error(err, "Failed to get source docker secret")
		r.statusUpdate(ctx, &now, &dockerCredCR, nil, nil, metav1.Condition{
			Type:               "Ready",
			Status:             metav1.ConditionFalse,
			Reason:             "SourceSecretNotFound",
			Message:            err.Error(),
			LastTransitionTime: now,
		})
		return ctrl.Result{}, err
	}

	// Get all namespaces
	namespaceList := &corev1.NamespaceList{}
	if err := r.List(ctx, namespaceList); err != nil {
		log.Error(err, "Unable to list namespaces")
		return ctrl.Result{}, err
	}

	var synced []string
	var failed []string

	// Get Docker Cred Secrete from source namespace
	targetSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: dockerCredCR.Spec.SourceNamespace,
			Name:      dockerCredCR.Spec.SourceSecretName,
		},
	}

	for _, ns := range namespaceList.Items {
		// Check if it is a targeted namespace
		if !strings.HasPrefix(ns.Name, dockerCredCR.Spec.TargetNamespacePrefix) {
			continue
		}

		// Sync Secret
		_, err := controllerutil.CreateOrUpdate(ctx, r.Client, targetSecret, func() error {
			targetSecret.Type = sourceSecret.Type
			targetSecret.Data = sourceSecret.Data

			if targetSecret.Annotations == nil {
				targetSecret.Annotations = map[string]string{}
			}
			targetSecret.Annotations["dax.vo/managed-by"] = dockerCredCR.Name
			return nil
		})

		if err != nil {
			log.Error(err, "failed to sync secret to namespace", "namespace", ns.Name)
			failed = append(failed, ns.Name)
		} else {
			synced = append(synced, ns.Name)
		}
	}

	r.statusUpdate(ctx, &now, &dockerCredCR, synced, failed, metav1.Condition{
		Type:   "Ready",
		Status: metav1.ConditionTrue,
		Reason: "AllSecretsSynced",
		Message: fmt.Sprintf(
			"Synced to %d namespaces, failed in %d",
			len(synced),
			len(failed),
		),
		LastTransitionTime: now,
	})

	// Config reconciling loop period
	refreshInterval := time.Duration(dockerCredCR.Spec.RefreshIntervalSeconds) * time.Second
	log.Info("Reconciled Docker Credential Sync")

	return ctrl.Result{RequeueAfter: refreshInterval}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DockerCredentialSyncReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&daxvov1.DockerCredentialSync{}).
		Named("dockercredentialsync").
		Owns(&corev1.Secret{}).
		Complete(r)
}

// Update status helper function
func (r *DockerCredentialSyncReconciler) statusUpdate(
	ctx context.Context,
	now *metav1.Time,
	dc *daxvov1.DockerCredentialSync,
	synced []string,
	failed []string,
	condition metav1.Condition,
) {
	dc.Status.LastSyncedTime = now
	dc.Status.SyncedNamespaces = synced
	dc.Status.FailedNamespaces = failed
	dc.Status.Conditions = condition

	if err := r.Status().Update(ctx, dc); err != nil {
		ctrl.Log.Error(err, "failed to update status")
	}
}
