package admit

import (
  "errors"
  "log"
  "strings"
  "encoding/json"
  "io/ioutil"
  "net/http"
  "strings"

  "github.com/nokia/danm/pkg/cnidel"
  admissionv1 "k8s.io/api/admission/v1"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  "k8s.io/apimachinery/pkg/runtime"
  "k8s.io/apimachinery/pkg/runtime/serializer"
  "github.com/nokia/danm/pkg/cnidel"
)

type Patch struct {
  Op    string          `json:"op"`
  Path  string          `json:"path"`
  Value interface{}     `json:"value,omitempty"`
}

func DecodeAdmissionReview(httpRequest *http.Request) (admissionv1.AdmissionReview,error) {
  var payload []byte
  reviewRequest := admissionv1.AdmissionReview{}
  if httpRequest.Body == nil {
    return reviewRequest, errors.New("Received review request is empty!")
  }
  payload, err := ioutil.ReadAll(httpRequest.Body);
  if err != nil {
    return reviewRequest, err
  }
  codecs := serializer.NewCodecFactory(runtime.NewScheme())
  deserializer := codecs.UniversalDeserializer()
  _, _, err = deserializer.Decode(payload, nil, &reviewRequest)
  return reviewRequest, err
}

func SendErroneousAdmissionResponse(responseWriter http.ResponseWriter, request *admissionv1.AdmissionRequest, err error) {
  log.Println("ERROR: Admitting resource failed with error:" + err.Error())
  failedResponse := &admissionv1.AdmissionResponse {
    Result: &metav1.Status {
      Message: err.Error(),
    },
    Allowed: false,
  }
  if request != nil {
    failedResponse.UID = request.UID
  }
  responseAdmissionReview := admissionv1.AdmissionReview {
    Response: failedResponse,
  }
  SendAdmissionResponse(responseWriter, responseAdmissionReview)
}

func SendAdmissionResponse(responseWriter http.ResponseWriter, reviewResponse admissionv1.AdmissionReview) {
  respBytes, err := json.Marshal(reviewResponse)
  if err != nil {
    log.Println("ERROR: Failed to send AdmissionResponse for request:" + string(reviewResponse.Response.UID) + " because JSON marshalling failed with error:" + err.Error())
  }
  responseWriter.Header().Set("Content-Type", "application/json")
  _, err = responseWriter.Write(respBytes)
  if err != nil {
    log.Println("ERROR: Failed to send AdmissionRespons for request:" + string(reviewResponse.Response.UID) + " because putting the HTTP response on the wire failed with error:" + err.Error())
  }
}

func CreateReviewResponseFromPatches(patchList []Patch) *admissionv1.AdmissionResponse {
  reviewResponse := admissionv1.AdmissionResponse{Allowed: true}
  var patches []byte
  var err error
  if len(patchList) > 0 {
    patches, err = json.Marshal(patchList)
    if err != nil {
      reviewResponse.Allowed = false
      reviewResponse.Result  = &metav1.Status{ Message: "List of patches could not be encoded, because:" + err.Error(),}
      return &reviewResponse
    }
  }
  if len(patches) > 0 {
    reviewResponse.Patch = patches
    pt := admissionv1.PatchTypeJSONPatch
    reviewResponse.PatchType = &pt
  }
  return &reviewResponse
}

func CreateGenericPatchFromChange(path string, value interface{}) Patch {
  patch := Patch {
    Op:    "replace",
    Path:  path,
    Value: value,
  }
  return patch
}

func IsTypeDynamic(cniType string) bool {
  neType := strings.ToLower(cniType)
  if _, ok := cnidel.SupportedNativeCnis[neType]; ok || neType == "" || neType == "ipvlan" {
    return true
  }
  return false
}
