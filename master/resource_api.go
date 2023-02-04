package master

import (
	"context"
	"errors"

	proto "github.com/a11en4sec/crawler/proto/crawler"
	"github.com/golang/protobuf/ptypes/empty"
)

func (m *Master) DeleteResource(ctx context.Context, spec *proto.ResourceSpec, empty *empty.Empty) error {
	r, ok := m.resources[spec.Name]

	if !ok {
		return errors.New("no such task")
	}

	if _, err := m.etcdCli.Delete(context.Background(), getResourcePath(spec.Name)); err != nil {
		return err
	}

	if r.AssignedNode != "" {
		nodeID, err := getNodeID(r.AssignedNode)
		if err != nil {
			return err
		}

		if ns, ok := m.workNodes[nodeID]; ok {
			ns.Payload -= 1
		}
	}
	return nil
}

func (m *Master) AddResource(ctx context.Context, req *proto.ResourceSpec, resp *proto.NodeSpec) error {
	nodeSpec, err := m.addResources(&ResourceSpec{Name: req.Name})
	if nodeSpec != nil {
		resp.Id = nodeSpec.Node.Id
		resp.Address = nodeSpec.Node.Address
	}
	return err
}
