package client

import (
	"testing"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/golang/mock/gomock"
)

var (
	t          *testing.T
	domainName = "user.mshneorson"
)

func PrepareMockClient(_t *testing.T) {
	t = _t
}

func AccTestZmsClient() (*MockZmsClient, error) {
	mockCtrl := gomock.NewController(t)
	clientMock := NewMockZmsClient(mockCtrl)
	clientMock.EXPECT().GetRole(domainName, "foo123").Return(simpleRole(), nil).AnyTimes()
	clientMock.EXPECT().GetRole(domainName, "foo123").Return(nil, nil)
	clientMock.EXPECT().PutRole(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	return clientMock, nil
}

func simpleRole() *zms.Role {
	return &zms.Role{Name: "test"}
}
