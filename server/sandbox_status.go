package server

import (
	"time"

	"github.com/cri-o/cri-o/oci"
	"github.com/cri-o/cri-o/version"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	pb "k8s.io/kubernetes/pkg/kubelet/apis/cri/runtime/v1alpha2"
)

// PodSandboxStatus returns the Status of the PodSandbox.
func (s *Server) PodSandboxStatus(ctx context.Context, req *pb.PodSandboxStatusRequest) (resp *pb.PodSandboxStatusResponse, err error) {
	const operation = "pod_sandbox_status"
	defer func() {
		recordOperation(operation, time.Now())
		recordError(operation, err)
	}()

	logrus.Debugf("PodSandboxStatusRequest %+v", req)
	sb, err := s.getPodSandboxFromRequest(req.PodSandboxId)
	if err != nil {
		return nil, err
	}

	podInfraContainer := sb.InfraContainer()
	cState := s.Runtime().ContainerStatus(podInfraContainer)

	rStatus := pb.PodSandboxState_SANDBOX_NOTREADY
	if cState.Status == oci.ContainerStateRunning {
		rStatus = pb.PodSandboxState_SANDBOX_READY
	}

	linux := &pb.LinuxPodSandboxStatus{
		Namespaces: &pb.Namespace{
			Options: &pb.NamespaceOption{
				Network: sb.NamespaceOptions().GetNetwork(),
				Ipc:     sb.NamespaceOptions().GetIpc(),
				Pid:     sb.NamespaceOptions().GetPid(),
			},
		},
	}

	sandboxID := sb.ID()
	resp = &pb.PodSandboxStatusResponse{
		Status: &pb.PodSandboxStatus{
			Id:          sandboxID,
			CreatedAt:   podInfraContainer.CreatedAt().UnixNano(),
			Network:     &pb.PodSandboxNetworkStatus{Ip: sb.IP()},
			State:       rStatus,
			Labels:      sb.Labels(),
			Annotations: sb.Annotations(),
			Metadata:    sb.Metadata(),
			Linux:       linux,
		},
	}

	if req.Verbose {
		resp.Info = map[string]string{
			"version": version.Version,
		}
	}

	logrus.Debugf("PodSandboxStatusResponse: %+v", resp)
	return resp, nil
}
