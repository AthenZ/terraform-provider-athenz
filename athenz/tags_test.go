package athenz

import (
	"math/rand"
	"testing"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/stretchr/testify/assert"
	ast "gotest.tools/assert"
)

func Test_allElementsNotContained(t *testing.T) {

	ast.DeepEqual(t, allElementsNotContained([]string{"s1", "s2", "s3"}, []string{"s1", "s2"}), []string{"s3"})
	ast.DeepEqual(t, allElementsNotContained([]string{}, []string{"s1", "s2"}), []string{})
	ast.DeepEqual(t, allElementsNotContained([]string{"s1", "s2", "s3"}, []string{}), []string{"s1", "s2", "s3"})
	ast.DeepEqual(t, allElementsNotContained([]string{"s1", "s2", "s3"}, []string{"s4", "s5"}), []string{"s1", "s2", "s3"})
	ast.DeepEqual(t, allElementsNotContained([]string{"s1", "s2", "s3"}, []string{"s1", "s2", "s3"}), []string{})
}

func Test_convertTagComponentValueListToStringList(t *testing.T) {
	ast.DeepEqual(t,
		convertTagComponentValueListToStringList([]zms.TagCompoundValue{"s1", "s2", "s3"}),
		[]string{"s1", "s2", "s3"})
	ast.DeepEqual(t, convertTagComponentValueListToStringList([]zms.TagCompoundValue{}), []string{})
}

func Test_makeTagsValue(t *testing.T) {
	ast.DeepEqual(t, makeTagsValue("s1,s2,s3"),
		&zms.TagValueList{List: []zms.TagCompoundValue{"s1", "s2", "s3"}})
	ast.DeepEqual(t, makeTagsValue(""), &zms.TagValueList{List: []zms.TagCompoundValue{}})
}

//works if changing in line 59 from .(*schema.Set).List() to .([]interface)
func Test_expandRoleTags(t *testing.T) {
	actual := expandRoleTags(map[string]interface{}{"key1": "v1k1,v2k1", "key2": "v1k2,v2k2,v3k2"})
	expected := buildMapForSchemaTest([]string{"key1", "key2"}, []int{2, 3}, []string{"v1k1", "v2k1", "v1k2", "v2k2", "v3k2"})
	checkMapEquals(t, expected, actual)
	actual = expandRoleTags(map[string]interface{}{"key1": "", "key2": "v1k2,v2k2,v3k2"})
	expected = buildMapForSchemaTest([]string{"key2"}, []int{3}, []string{"v1k2", "v2k2", "v3k2"})
	checkMapEquals(t, expected, actual)
	actual = expandRoleTags(map[string]interface{}{"key1": "v1k2,v2k2,v3k2", "key2": ""})
	expected = buildMapForSchemaTest([]string{"key1"}, []int{3}, []string{"v1k2", "v2k2", "v3k2"})
	checkMapEquals(t, expected, actual)
	actual = expandRoleTags(map[string]interface{}{"key1": "v1k2,v2k2,v3k2", "key2": "yy,yy1", "key3": "zz1,zz2"})
	expected = buildMapForSchemaTest([]string{"key1", "key2", "key3"}, []int{3, 2, 2}, []string{"v1k2", "v2k2", "v3k2", "yy", "yy1", "zz1", "zz2"})
	checkMapEquals(t, expected, actual)
}

func checkMapEquals(t *testing.T, expected map[zms.CompoundName]*zms.TagValueList, actual map[zms.CompoundName]*zms.TagValueList) {
	for key, val := range expected {
		assert.ElementsMatch(t, (*val).List, (*actual[key]).List)
	}
}

func Test_flattenTag(t *testing.T) {
	assert.EqualValues(t,
		map[string]interface{}{
			"key1": "v1k1,v2k1",
			"key2": "v1k2,v2k2,v3k2"},
		flattenTag(buildMapForSchemaTest([]string{"key1", "key2"}, []int{2, 3}, []string{"v1k1", "v2k1", "v1k2", "v2k2", "v3k2"})))

	assert.EqualValues(t,
		map[string]interface{}{"key2": "v1k2,v2k2,v3k2"},
		flattenTag(buildMapForSchemaTest([]string{"key1", "key2"}, []int{0, 3}, []string{"v1k2", "v2k2", "v3k2"})))
}

func buildMapForSchemaTest(keys []string, sizes []int, val []string) map[zms.CompoundName]*zms.TagValueList {
	finalMap := map[zms.CompoundName]*zms.TagValueList{}
	finalValues := makeZmsTagValueList(sizes, val)
	for i, _ := range keys {
		finalMap[zms.CompoundName(keys[i])] = finalValues[i]
	}
	return finalMap
}
func makeZmsTagValueList(sizes []int, list []string) []*zms.TagValueList {
	finalArr := []*zms.TagValueList{}
	count := 0
	for _, index := range sizes {
		valuesArr := []zms.TagCompoundValue{}
		for i := count; i < count+index; i++ {
			valuesArr = append(valuesArr, zms.TagCompoundValue(list[i]))
		}
		count += index
		finalArr = append(finalArr, &zms.TagValueList{List: valuesArr})
	}
	return finalArr

}

// map init = keys - [arr of keys], sizes - [size of arr values], values - [all the values of all the key chained]
func makeTagsSchemaArr(keys []string, sizes []int, values []string) []interface{} {
	valLocation := 0
	count := 0
	finalArr := []interface{}{}

	for _, key := range keys {
		newMap := map[string]interface{}{}
		newArr := []interface{}{}
		for i := count; i < count+sizes[valLocation]; i++ {
			newArr = append(newArr, values[i])
		}
		count = count + sizes[valLocation]
		valLocation += 1
		newMap["key"] = key
		//newMap["values"] = newArr

		f := schema.SchemaSetFunc(func(interface{}) int { return rand.Int() })
		newMap["values"] = schema.NewSet(f, newArr)

		finalArr = append(finalArr, newMap)
	}
	return finalArr
}
