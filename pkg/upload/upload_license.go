package upload

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"github.com/replicatedhq/kots/pkg/auth"
	"github.com/replicatedhq/kots/pkg/k8sutil"
	"github.com/replicatedhq/kots/pkg/logger"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type UploadLicenseOptions struct {
	Namespace             string
	KubernetesConfigFlags *genericclioptions.ConfigFlags
	NewAppName            string
}

func UploadLicense(path string, uploadLicenseOptions UploadLicenseOptions) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Wrap(err, "failed to read license file")
	}
	license := string(b)

	// Make sure we have a name or slug
	if uploadLicenseOptions.NewAppName == "" {
		appName, err := relentlesslyPromptForAppName("")
		if err != nil {
			return errors.Wrap(err, "failed to prompt for app name")
		}

		uploadLicenseOptions.NewAppName = appName
	}

	// Find the kotadm-api pod
	log := logger.NewLogger()
	log.ActionWithSpinner("Uploading license to Admin Console")

	clientset, err := k8sutil.GetClientset(uploadLicenseOptions.KubernetesConfigFlags)
	if err != nil {
		log.FinishSpinnerWithError()
		return errors.Wrap(err, "failed to get clisnetset")
	}

	podName, err := k8sutil.FindKotsadm(clientset, uploadLicenseOptions.Namespace)
	if err != nil {
		log.FinishSpinnerWithError()
		return errors.Wrap(err, "failed to find kotsadm pod")
	}

	stopCh := make(chan struct{})
	defer close(stopCh)

	_, errChan, err := k8sutil.PortForward(uploadLicenseOptions.KubernetesConfigFlags, 3000, 3000, uploadLicenseOptions.Namespace, podName, false, stopCh, log)
	if err != nil {
		log.FinishSpinnerWithError()
		return errors.Wrap(err, "failed to start port forwarding")
	}

	go func() {
		select {
		case err := <-errChan:
			if err != nil {
				log.Error(err)
			}
		case <-stopCh:
		}
	}()

	// upload using http to the pod directly
	req, err := createUploadLicenseRequest(license, uploadLicenseOptions, "http://localhost:3000/api/v1/kots/license")
	if err != nil {
		log.FinishSpinnerWithError()
		return errors.Wrap(err, "failed to create upload request")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.FinishSpinnerWithError()
		return errors.Wrap(err, "failed to execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		log.FinishSpinnerWithError()
		return errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.FinishSpinnerWithError()
		return errors.Wrap(err, "failed to read response body")
	}
	type UploadResponse struct {
		URI string `json:"uri"`
	}
	var uploadResponse UploadResponse
	if err := json.Unmarshal(respBody, &uploadResponse); err != nil {
		log.FinishSpinnerWithError()
		return errors.Wrap(err, "failed to unmarshal response")
	}

	log.FinishSpinner()

	return nil
}

func createUploadLicenseRequest(license string, uploadLicenseOptions UploadLicenseOptions, uri string) (*http.Request, error) {
	body := map[string]string{
		"name":    uploadLicenseOptions.NewAppName,
		"license": license,
	}

	b, err := json.Marshal(body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal json")
	}

	authSlug, err := auth.GetOrCreateAuthSlug(uploadLicenseOptions.KubernetesConfigFlags, uploadLicenseOptions.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get kotsadm auth slug")
	}

	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(b))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authSlug)
	return req, nil
}
