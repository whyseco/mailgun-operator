package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MailgunRouteSpec defines the desired state of MailgunRoute
// +k8s:openapi-gen=true
type MailgunRouteSpec struct {
	// Domain to create in mailgun: https://help.mailgun.com/hc/en-us/articles/202256730-How-Do-I-Pick-a-Domain-Name-for-My-Mailgun-Account-
	Domain string `json:"domain"`
	// API key to authenticate to mailgun API https://help.mailgun.com/hc/en-us/articles/203380100-Where-Can-I-Find-My-API-Key-and-SMTP-Credentials-
	ApiKey string `json:"apiKey"`
	// See https://documentation.mailgun.com/en/latest/api-routes.html#routes
	Expression  string `json:"expression"`
	Description string `json:"description,omitempty"`
	Priority    int    `json:"priority"`
	// +kubebuilder:validation:MinItems=1
	// +listType=set
	Actions []string `json:"actions,omitempty"`
}

// MailgunRouteStatus defines the observed state of MailgunRoute
// +k8s:openapi-gen=true
type MailgunRouteStatus struct {
	Id string `json:"id"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MailgunRoute is the Schema for the mailgunroutes API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=mailgunroutes,scope=Namespaced
type MailgunRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MailgunRouteSpec   `json:"spec,omitempty"`
	Status MailgunRouteStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MailgunRouteList contains a list of MailgunRoute
type MailgunRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MailgunRoute `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MailgunRoute{}, &MailgunRouteList{})
}
