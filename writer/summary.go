package writer

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"github.com/dineshba/tf-summarize/terraformstate"
)

type Resource struct {
	Tag    string `json:"tag"`
	Count  int    `json:"Count"`
	Action string `json:"action"`
}

type SummaryeWriter struct {
	changes map[string]terraformstate.ResourceChanges
}

func findDifferences(resource, updatedResource map[string]interface{}) map[string]interface{} {
	differences := make(map[string]interface{})
	for key, value1 := range resource {
		if value2, ok := updatedResource[key]; ok {
			if !reflect.DeepEqual(value1, value2) {
				differences[key] = value2
			}
		} else {
			differences[key] = nil
		}
	}

	for key, value2 := range updatedResource {
		if _, ok := updatedResource[key]; !ok {
			differences[key] = value2
		}
	}

	return differences
}

func adjustResourcesArray(resourcesArr []Resource, resource Resource) []Resource {
	for i, r := range resourcesArr {
		if r.Tag == resource.Tag {
			resourcesArr[i].Count = resourcesArr[i].Count + 1
			return resourcesArr
		}
	}
	return append(resourcesArr, resource)
}

func (t SummaryeWriter) Write(writer io.Writer) error {

	var resourcesSummary []Resource

	for _, change := range tableOrder {
		changedResources := t.changes[change]

		for _, changedResource := range changedResources {
			if change == "update" {
				var afterChange map[string]interface{}
				var beforeChange map[string]interface{}
				err := json.Unmarshal(changedResource.Change.After, &afterChange)
				if err != nil {
					fmt.Println("Error:", err)
				}
				err = json.Unmarshal(changedResource.Change.Before, &beforeChange)
				if err != nil {
					fmt.Println("Error:", err)

				}
				differences := findDifferences(beforeChange, afterChange)
				if len(differences) > 0 {
					for key := range differences {
						resource := Resource{Count: 1, Action: change, Tag: fmt.Sprintf("%s.%s", changedResource.Type, key)}
						resourcesSummary = adjustResourcesArray(resourcesSummary, resource)
					}
				}
			} else {
				resource := Resource{Count: 1, Action: change, Tag: changedResource.Type}
				resourcesSummary = adjustResourcesArray(resourcesSummary, resource)
			}
		}
	}
	for _, summary := range resourcesSummary {
		if summary.Count > 1 {
			fmt.Printf("%s - %ss %d times\n", summary.Tag, summary.Action, summary.Count)
		} else {
			fmt.Printf("%s - %ss %d time\n", summary.Tag, summary.Action, summary.Count)
		}
	}
	return nil
}

func NewSummaryWriter(changes map[string]terraformstate.ResourceChanges) Writer {
	return SummaryeWriter{
		changes: changes,
	}
}
