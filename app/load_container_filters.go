package app

import (
        "os"
        "fmt"
        "github.com/Jeffail/gabs"
        "github.com/weaveworks/scope/render"
)

func getContainerTopologyOptions() ([]APITopologyOption, error) {
        var toptions []APITopologyOption

        // get JSON string from environment variable
        s := os.Getenv("CONTAINER_FILTERS")

        arr, err := gabs.ParseJSON([]byte(s))
        if err != nil {
                fmt.Println(err)
                return toptions, err
        }
        filters, err := arr.Children()
        if err != nil {
                fmt.Println(err)
                return toptions, err
        }
        for _, f := range filters {
                labels, _ := f.S("acceptedLabels").Children()
                firstLabel := labels[0]
                sl := firstLabel.Data().(string)
                v := APITopologyOption{Value:f.Path("filterId").Data().(string), Label:f.Path("filterTitle").Data().(string), filter:render.IsDesired(sl), filterPseudo:false, filterLabel:sl}
                toptions = append(toptions,v)
        }
        if err != nil {
                fmt.Println(err)
                return toptions, err
        }
        
        // Add option to view all
        all := APITopologyOption{Value:"all", Label:"All", filter:nil, filterPseudo:false, filterLabel:""}
        toptions = append(toptions,all)
        return toptions, nil
}
