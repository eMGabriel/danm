package admit_tests

import (
  "strings"
  "testing"
  "encoding/json"
  danmtypes "github.com/nokia/danm/crd/apis/danm/v1"
  "github.com/nokia/danm/pkg/admit"
  httpstub "github.com/nokia/danm/test/stubs/http"
  "github.com/nokia/danm/test/utils"
  admissionv1 "k8s.io/api/admission/v1"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
  validateConfs = []danmtypes.TenantConfig {
    {
      ObjectMeta: metav1.ObjectMeta{Name: "malformed"},
      HostDevices: []danmtypes.IfaceProfile{
        {Name: "ens4", VniType: "vxlan", VniRange: "700-710", Alloc: utils.AllocFor5k},
      },
    },
    {
      ObjectMeta: metav1.ObjectMeta{Name: "invalid-type"},
      TypeMeta:   metav1.TypeMeta{Kind: "invalid"},
      HostDevices: []danmtypes.IfaceProfile{
        {Name: "ens4", VniType: "vxlan", VniRange: "700-710", Alloc: utils.AllocFor5k},
      },
    },
    {
      ObjectMeta: metav1.ObjectMeta{Name: "empty-config"}, TypeMeta: metav1.TypeMeta{Kind: "TenantConfig"},
    },
    {
      ObjectMeta: metav1.ObjectMeta{Name: "noname"}, TypeMeta: metav1.TypeMeta{Kind: "TenantConfig"},
      HostDevices: []danmtypes.IfaceProfile{
        {Name: "ens4", VniType: "vxlan", VniRange: "700-710"},
        {VniType: "vlan", VniRange: "200,500-510"},
      },
    },
    {
      ObjectMeta: metav1.ObjectMeta{Name: "norange"}, TypeMeta: metav1.TypeMeta{Kind: "TenantConfig"},
      HostDevices: []danmtypes.IfaceProfile{
        {Name: "ens4", VniType: "vxlan"},
        {Name: "ens5", VniType: "vlan", VniRange: "700-710"},
      },
    },
    {
      ObjectMeta: metav1.ObjectMeta{Name: "notype"}, TypeMeta: metav1.TypeMeta{Kind: "TenantConfig"},
      HostDevices: []danmtypes.IfaceProfile{
        {Name: "ens4", VniType: "vxlan", VniRange: "700-710"},
        {Name: "ens5", VniRange: "700-710"},
      },
    },
    {
      ObjectMeta: metav1.ObjectMeta{Name: "invalid-vni-type"}, TypeMeta: metav1.TypeMeta{Kind: "TenantConfig"},
      HostDevices: []danmtypes.IfaceProfile{
        {Name: "ens4", VniType: "vxlan2", VniRange: "700-710"},
      },
    },
    {
      ObjectMeta: metav1.ObjectMeta{Name: "invalid-vni-value"}, TypeMeta: metav1.TypeMeta{Kind: "TenantConfig"},
      HostDevices: []danmtypes.IfaceProfile{
        {Name: "ens4", VniType: "vxlan", VniRange: "700-71a0"},
      },
    },
    {
      ObjectMeta: metav1.ObjectMeta{Name: "invalid-vni-range"}, TypeMeta: metav1.TypeMeta{Kind: "TenantConfig"},
      HostDevices: []danmtypes.IfaceProfile{
        {Name: "ens4", VniType: "vxlan", VniRange: "900-4999,5001"},
      },
    },
    {
      ObjectMeta: metav1.ObjectMeta{Name: "valid-vni-range"}, TypeMeta: metav1.TypeMeta{Kind: "TenantConfig"},
      HostDevices: []danmtypes.IfaceProfile{
        {Name: "ens4", VniType: "vxlan", VniRange: "900-4999,5000"},
      },
    },
    {
      ObjectMeta: metav1.ObjectMeta{Name: "manual-alloc-old"}, TypeMeta: metav1.TypeMeta{Kind: "TenantConfig"},
      HostDevices: []danmtypes.IfaceProfile{
        {Name: "ens4", VniType: "vxlan", VniRange: "700-710", Alloc: utils.AllocFor5k},
      },
    },
    {
      ObjectMeta: metav1.ObjectMeta{Name: "manual-alloc"}, TypeMeta: metav1.TypeMeta{Kind: "TenantConfig"},
      HostDevices: []danmtypes.IfaceProfile{
        {Name: "ens4", VniType: "vxlan", VniRange: "700-710", Alloc: utils.AllocFor5k},
        {Name: "nokia.k8s.io/sriov_ens1f0", VniType: "vlan", VniRange: "700-710"},
      },
    },
    {
      ObjectMeta: metav1.ObjectMeta{Name: "nonetype"}, TypeMeta: metav1.TypeMeta{Kind: "TenantConfig"},
      NetworkIds: map[string]string{
        "": "asd",
      },
    },
    {
      ObjectMeta: metav1.ObjectMeta{Name: "nonid"}, TypeMeta: metav1.TypeMeta{Kind: "TenantConfig"},
      NetworkIds: map[string]string{
        "flannel": "",
      },
    },
    {
      ObjectMeta: metav1.ObjectMeta{Name: "longnid"}, TypeMeta: metav1.TypeMeta{Kind: "TenantConfig"},
      NetworkIds: map[string]string{
        "flannel": "abcdefghijkl",
      },
    },
    {
      ObjectMeta: metav1.ObjectMeta{Name: "longnid-sriov"}, TypeMeta: metav1.TypeMeta{Kind: "TenantConfig"},
      NetworkIds: map[string]string{
        "flannel": "abcdefghijkl",
        "sriov":   "abcdefghijkl",
      },
    },
    {
      ObjectMeta: metav1.ObjectMeta{Name: "shortnid"}, TypeMeta: metav1.TypeMeta{Kind: "TenantConfig"},
      NetworkIds: map[string]string{
        "flannel": "abcdefghijk",
        "sriov":   "abcdefghij",
      },
    },
    {
      ObjectMeta: metav1.ObjectMeta{Name: "old-iface"}, TypeMeta: metav1.TypeMeta{Kind: "TenantConfig"},
      HostDevices: []danmtypes.IfaceProfile{
        {Name: "ens4", VniType: "vxlan", VniRange: "900-4999,5000", Alloc: utils.AllocFor5k},
      },
    },
    {
      ObjectMeta: metav1.ObjectMeta{Name: "new-iface"}, TypeMeta: metav1.TypeMeta{Kind: "TenantConfig"},
      HostDevices: []danmtypes.IfaceProfile{
        {Name: "ens4", VniType: "vxlan", VniRange: "900-4999,5000", Alloc: utils.AllocFor5k},
      },
      NetworkIds: map[string]string{
        "flannel": "flannel",
      },
    },
  }
)

var validateTconfTcs = []struct {
  tcName          string
  oldTconfName    string
  newTconfName    string
  opType          admissionv1.Operation
  isErrorExpected bool
  expectedPatches []admit.Patch
}{
  {"emptyRequest", "", "", "", true, nil},
  {"malformedOldObject", "malformed", "", "", true, nil},
  {"malformedNewObject", "", "malformed", "", true, nil},
  {"objectWithInvalidType", "", "invalid-type", "", true, nil},
  {"emptyCofig", "", "empty-config", "", true, nil},
  {"interfaceProfileWithoutName", "", "noname", "", true, nil},
  {"interfaceProfileWithoutVniRange", "", "norange", "", true, nil},
  {"interfaceProfileWithoutVniType", "", "notype", "", true, nil},
  {"interfaceProfileWithInvalidVniType", "", "invalid-vni-type", "", true, nil},
  {"interfaceProfileWithInvalidVniValue", "", "invalid-vni-value", "", true, nil},
  {"interfaceProfileWithInvalidVniRange", "", "invalid-vni-range", "", true, nil},
  {"interfaceProfileWithValidVniRange", "", "valid-vni-range", "", false, expectedPatch},
  {"interfaceProfileWithSetAlloc", "", "manual-alloc", admissionv1.Create, true, nil},
  {"interfaceProfileChangeWithAlloc", "manual-alloc-old", "manual-alloc", admissionv1.Update, false, expectedPatch},
  {"networkIdWithoutKey", "", "nonid", "", true, nil},
  {"networkIdWithoutValue", "", "nonetype", "", true, nil},
  {"longNidWithStaticNeType", "", "longnid", "", false, nil},
  {"longNidWithDynamicNeType", "", "longnid-sriov", "", true, nil},
  {"okayNids", "", "shortnid", "", false, nil},
  {"noChangeInIfaces", "old-iface", "new-iface", admissionv1.Update, false, nil},
}

var (
  expectedPatch = []admit.Patch {
    {Path: "/hostDevices"},
  }
)

func TestValidateTenantConfig(t *testing.T) {
  validator := admit.Validator{}
  for _, tc := range validateTconfTcs {
    t.Run(tc.tcName, func(t *testing.T) {
      writerStub := httpstub.NewWriterStub()
      oldTconf, shouldOldMalform := getTestConf(tc.oldTconfName, validateConfs)
      newTconf, shouldNewMalform := getTestConf(tc.newTconfName, validateConfs)
      request,err := utils.CreateHttpRequest(oldTconf, newTconf, shouldOldMalform, shouldNewMalform, tc.opType)
      if err != nil {
        t.Errorf("Could not create test HTTP Request object, because:%v", err)
        return
      }
      validator.ValidateTenantConfig(writerStub, request)
      err = utils.ValidateHttpResponse(writerStub, tc.isErrorExpected, tc.expectedPatches)
      if err != nil {
        t.Errorf("Received HTTP Response did not match expectation, because:%v", err)
        return
      }
    })
  }
}

func getTestConf(name string, confs []danmtypes.TenantConfig) ([]byte, bool) {
  tconf := utils.GetTconf(name, confs)
  if tconf == nil {
    return nil, false
  }
  var shouldItMalform bool
  if strings.HasPrefix(tconf.ObjectMeta.Name, "malform") {
    shouldItMalform = true
  }
  tconfBinary,_ := json.Marshal(tconf)
  return tconfBinary, shouldItMalform
}
