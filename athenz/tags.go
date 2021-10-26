package athenz

import (
	"strings"

	"github.com/AthenZ/athenz/clients/go/zms"
)

// flattenTag - takes the tag form the zms and return a tag schema
func flattenTag(tagsMap map[zms.CompoundName]*zms.TagValueList) map[string]interface{} {
	tags := map[string]interface{}{}
	for tagKey, valueList := range tagsMap {
		tagValue := ""
		for _, m := range valueList.List {
			if tagValue != "" {
				tagValue = tagValue + "," + string(m)
			} else {
				tagValue = string(m)
			}
		}
		if tagValue != "" {
			tags[string(tagKey)] = tagValue
		}
	}
	return tags
}

func expandRoleTags(tagsMap map[string]interface{}) map[zms.CompoundName]*zms.TagValueList {
	return expandTagsMap(tagsMap)
}
func expandTagsMap(tagsMap map[string]interface{}) map[zms.CompoundName]*zms.TagValueList {
	roleTags := map[zms.CompoundName]*zms.TagValueList{}
	for key, listVal := range tagsMap {
		tags := makeTagsValue(listVal.(string))
		if len(tags.List) > 0 {
			roleTags[zms.CompoundName(key)] = tags
		}
	}
	if len(roleTags) > 0 {
		return roleTags
	}
	return map[zms.CompoundName]*zms.TagValueList{zms.CompoundName("key"): &zms.TagValueList{List: []zms.TagCompoundValue{}}}
}

func makeTagsValue(tagsValues string) *zms.TagValueList {
	tagsList := make([]zms.TagCompoundValue, 0, len(tagsValues))
	tags := strings.Split(tagsValues, ",")
	for _, val := range tags {
		if val != "" {
			tagsList = append(tagsList, zms.TagCompoundValue(val))
		}
	}
	return &zms.TagValueList{List: tagsList}
}

func convertTagComponentValueListToStringList(list []zms.TagCompoundValue) []string {
	finalList := []string{}
	for _, valueToRemove := range list {
		finalList = append(finalList, string(valueToRemove))
	}
	return finalList
}

//return all the values that exists in arr1 and not in arr2
func allElementsNotContained(arr1 []string, arr2 []string) []string {
	finalArr := []string{}
	check := false
	for _, val1 := range arr1 {
		for _, val2 := range arr2 {
			if val1 == val2 {
				check = true
				break
			}
		}
		if check == false {
			finalArr = append(finalArr, val1)
		}
		check = false
	}
	return finalArr
}
