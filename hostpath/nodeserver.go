/*
Copyright 2017 The Kubernetes Authors.

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

package hostpath

import (
	"os"

	"github.com/golang/glog"
	"golang.org/x/net/context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/kubernetes/pkg/util/mount"

	"github.com/kubernetes-csi/drivers/csi-common"
)

type nodeServer struct {
	*csicommon.DefaultNodeServer
}

func (ns *nodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	targetPath := req.GetTargetPath()
	notMnt, err := mount.New("").IsLikelyNotMountPoint(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(targetPath, 0750); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			notMnt = true
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	if !notMnt {
		return &csi.NodePublishVolumeResponse{}, nil
	}

	fsType := req.GetVolumeCapability().GetMount().GetFsType()

	deviceId := ""
	if req.GetPublishVolumeInfo() != nil {
		deviceId = req.GetPublishVolumeInfo()[deviceID]
	}

	readOnly := req.GetReadonly()
	volumeId := req.GetVolumeId()
	attrib := req.GetVolumeAttributes()
	mountFlags := req.GetVolumeCapability().GetMount().GetMountFlags()

	glog.V(4).Infof("target %v\nfstype %v\ndevice %v\nreadonly %v\nattributes %v\n mountflags %v\n",
		targetPath, fsType, deviceId, readOnly, volumeId, attrib, mountFlags)

	options := []string{"bind"}
	if readOnly {
		options = append(options, "ro")
	}
	mounter := mount.New("")
	path := provisionRoot + volumeId
	if err := mounter.Mount(path, targetPath, "", options); err != nil {
		return nil, err
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	return &csi.NodeUnpublishVolumeResponse{}, nil
}
