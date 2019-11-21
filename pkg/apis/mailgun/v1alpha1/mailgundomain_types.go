package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SpamActionType string

const (
	Disabled SpamActionType = "disabled"
	Block    SpamActionType = "block"
	Tag      SpamActionType = "tag"
)

type WebSchemeType string

const (
	HTTP  SpamActionType = "http"
	HTTPS SpamActionType = "https"
)

// MailgunDomainSpec defines the desired state of MailgunDomain
// +k8s:openapi-gen=true
type MailgunDomainSpec struct {
	// Domain to create in mailgun: https://help.mailgun.com/hc/en-us/articles/202256730-How-Do-I-Pick-a-Domain-Name-for-My-Mailgun-Account-
	Domain string `json:"domain"`
	// API key to authenticate to mailgun API https://help.mailgun.com/hc/en-us/articles/203380100-Where-Can-I-Find-My-API-Key-and-SMTP-Credentials-
	ApiKey string `json:"apiKey"`

	// See https://documentation.mailgun.com/en/latest/api-domains.html#domains
	Password           string         `json:"password,omitempty"`
	SpamAction         SpamActionType `json:"spamAction,omitempty"`
	Wildcard           bool           `json:"wildcard,omitempty"`
	ForceDKIMAuthority bool           `json:"force_dkim_authority,omitempty"`
	DKIMKeySize        int            `json:"dkim_key_size,omitempty"`
	// +kubebuilder:validation:MinItems=0
	// +listType=set
	IPS       []string      `json:"ips,omitempty"`
	WebScheme WebSchemeType `json:"webScheme,omitempty"`
}

// MailgunDomainDnsRecord defines the receiving and sending dns record provided by mailgun
// +k8s:openapi-gen=true
type MailgunDomainDnsRecord struct {
	RecordType string `json:"recordType"`
	Valid      string `json:"valid"`
	Priority   string `json:"priority,omitempty"`
	Value      string `json:"value"`
	Name       string `json:"name,omitempty"`
}

// MailgunDomainStatus defines the observed state of MailgunDomain
// +k8s:openapi-gen=true
type MailgunDomainStatus struct {
	SendingDnsRecord   []MailgunDomainDnsRecord `json:"sendingDnsRecord,omitempty"`
	ReceivingDnsRecord []MailgunDomainDnsRecord `json:"receivingDnsRecord,omitempty"`
	DomainState        string                   `json:"domainState"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MailgunDomain is the Schema for the mailgundomains API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=mailgundomains,scope=Namespaced
type MailgunDomain struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MailgunDomainSpec   `json:"spec,omitempty"`
	Status MailgunDomainStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MailgunDomainList contains a list of MailgunDomain
type MailgunDomainList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MailgunDomain `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MailgunDomain{}, &MailgunDomainList{})
}
