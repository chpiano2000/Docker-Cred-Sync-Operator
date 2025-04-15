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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DockerCredentialSyncSpec defines the desired state of DockerCredentialSync.
type DockerCredentialSyncSpec struct {
	SourceSecretName       string `json:"sourceSecretName"`
	SourceNamespace        string `json:"sourceNamespace"`
	TargetNamespacePrefix  string `json:"targetNamespacePrefix"`
	RefreshIntervalSeconds int    `json:"refreshIntervalSeconds"`
}

// DockerCredentialSyncStatus defines the observed state of DockerCredentialSync.
type DockerCredentialSyncStatus struct {
	// LastSyncedTime is the timestamp of the last successful or attempted sync
	LastSyncedTime *metav1.Time `json:"lastSyncedTime,omitempty"`

	// SyncedNamespaces is the list of namespaces where the credential has been successfully synced
	SyncedNamespaces []string `json:"syncedNamespaces,omitempty"`

	// FailedNamespaces is the list of namespaces where syncing failed
	FailedNamespaces []string `json:"failedNamespaces,omitempty"`

	// Conditions represent the latest available observations of the object's state
	Conditions metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DockerCredentialSync is the Schema for the dockercredentialsyncs API.
type DockerCredentialSync struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DockerCredentialSyncSpec   `json:"spec,omitempty"`
	Status DockerCredentialSyncStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DockerCredentialSyncList contains a list of DockerCredentialSync.
type DockerCredentialSyncList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DockerCredentialSync `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DockerCredentialSync{}, &DockerCredentialSyncList{})
}
