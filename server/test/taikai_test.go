package test

import (
	taikaiv1 "github.com/forgeutah/taikai/protos/gen/go/taikai/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/testing/protocmp"
	"testing"
	"time"
)

type TaikaiSuite struct {
	suite.Suite
}

func TestTaikaiSuite(t *testing.T) {
	suite.Run(t, new(TaikaiSuite))
}

func (s *TaikaiSuite) SetupSuite() {
	initializeApiClient(5 * time.Second)
	loadTestData(s.T())
}

func assertProtoEqualitySortById(t *testing.T, expected, actual interface{}, opts ...cmp.Option) {
	theOpts := []cmp.Option{
		cmpopts.SortSlices(func(x, y *taikaiv1.Hello) bool {
			return *x.Id < *y.Id
		}),
		protocmp.Transform(),
		protocmp.SortRepeated(func(x, y *taikaiv1.Hello) bool {
			return *x.Id < *y.Id
		}),
	}
	theOpts = append(theOpts, opts...)
	diff := cmp.Diff(expected, actual, theOpts...)
	require.Empty(t, diff)
}
