package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MailgunWebhookSpec defines the desired state of MailgunWebhook
// +k8s:openapi-gen=true
type MailgunWebhookSpec struct {
	// Domain to use in mailgun
	Domain string `json:"domain"`
	// secret name where we can find apiKey
	SecretName string `json:"secretName"`
	// +kubebuilder:validation:MaxItems=3
	// +kubebuilder:validation:MinItems=0
	// +listType=set
	Clicked []string `json:"clicked,omitempty"`
	// +kubebuilder:validation:MaxItems=3
	// +kubebuilder:validation:MinItems=0
	// +listType=set
	Complained []string `json:"complained,omitempty"`
	// +kubebuilder:validation:MaxItems=3
	// +kubebuilder:validation:MinItems=0
	// +listType=set
	Delivered []string `json:"delivered,omitempty"`
	// +kubebuilder:validation:MaxItems=3
	// +kubebuilder:validation:MinItems=0
	// +listType=set
	Opened []string `json:"opened,omitempty"`
	// +kubebuilder:validation:MaxItems=3
	// +kubebuilder:validation:MinItems=0
	// +listType=set
	PermanentFail []string `json:"permanentFail,omitempty"`
	// +kubebuilder:validation:MaxItems=3
	// +kubebuilder:validation:MinItems=0
	// +listType=set
	TemporaryFail []string `json:"temporaryFail,omitempty"`
	// +kubebuilder:validation:MaxItems=3
	// +kubebuilder:validation:MinItems=0
	// +listType=set
	Unsubscribed []string `json:"unsubscribed,omitempty"`
}

// MailgunWebhookStatus defines the observed state of MailgunWebhook
// +k8s:openapi-gen=true
type MailgunWebhookStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Ready bool `json:"ready"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MailgunWebhook is the Schema for the mailgunwebhooks API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=mailgunwebhooks,scope=Namespaced
type MailgunWebhook struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MailgunWebhookSpec   `json:"spec,omitempty"`
	Status MailgunWebhookStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MailgunWebhookList contains a list of MailgunWebhook
type MailgunWebhookList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MailgunWebhook `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MailgunWebhook{}, &MailgunWebhookList{})
}
