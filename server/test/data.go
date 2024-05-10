package test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	taikaiv1 "github.com/forgeutah/taikai/protos/gen/go/taikai/v1"
	"github.com/emirpasic/gods/sets/hashset"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

type TestData struct {
	Hellos   []*taikaiv1.Hello
	HelloIds *hashset.Set
}

var LoadedTestData = TestData{}

func loadTestData(t *testing.T) {
	deleteAllData(t)
	loadAllData(t)
}

func loadAllData(t *testing.T) {
	loadHellos(t)
}

func deleteAllData(t *testing.T) {
	deleteHellos(t)
}

func deleteHellos(t *testing.T) {
	hellos, err := ListHellos(1000, 0, "")
	require.NoError(t, err)
	ids := lo.Map(hellos, func(item *taikaiv1.Hello, index int) string {
		return lo.FromPtr(item.Id)
	})
	require.NoError(t, DeleteHellos(ids))
}

func loadHellos(t *testing.T) {
	hellos := CreateRandomNumHellos(t)
	LoadedTestData.Hellos = hellos
	LoadedTestData.HelloIds = hashset.New()
	for _, hello := range hellos {
		LoadedTestData.HelloIds.Add(*hello.Id)
	}
}

func CreateRandomNumHellos(t *testing.T) []*taikaiv1.Hello {
	return CreateHellos(t, gofakeit.Number(5, 10))
}

func CreateHellos(t *testing.T, num int) []*taikaiv1.Hello {
	var err error
	hellos := []*taikaiv1.Hello{}
	for i := 0; i < num; i++ {
		hello := createRandomHelloProto(t)
		hellos = append(hellos, hello)
	}
	hellos, err = UpsertHellos(hellos)
	require.NoError(t, err)
	return hellos
}

func createRandomHelloProto(t *testing.T) *taikaiv1.Hello {
	hello := &taikaiv1.Hello{}
	err := gofakeit.Struct(hello)
	require.NoError(t, err)
	return hello
}

func randomizeHellos(t *testing.T, hellos []*taikaiv1.Hello) []*taikaiv1.Hello {
	randomized := []*taikaiv1.Hello{}
	for _, hello := range hellos {
		random := createRandomHelloProto(t)
		random.Id = hello.Id
		random.UpdatedAt = hello.UpdatedAt
		random.CreatedAt = hello.CreatedAt
		randomized = append(randomized, random)
	}
	return randomized
}
