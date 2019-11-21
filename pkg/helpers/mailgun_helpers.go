package helpers

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetApiKey(ctx context.Context, reqLogger logr.Logger, client client.Client, secretName string, namespace string) (string, error) {
	secret := &corev1.Secret{}
	err := client.Get(ctx, types.NamespacedName{Name: secretName, Namespace: namespace}, secret)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Secret does not exist", "Namespace", namespace, "Name", secretName)
		return "", err
	}

	apiKey := string(secret.Data["apiKey"])

	return apiKey, nil
}
