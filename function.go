package gcpcleaner

import (
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"os"
	"log"
	"strings"
)

// PubSubMessage is the payload of a Pub/Sub event. Please refer to the docs for
// additional information regarding Pub/Sub events.
type PubSubMessage struct {
	Data []byte `json:"data"`
}

// Clean assumes that it is running inside GCP using a service account with the appropriate permissions.
func Clean(ctx context.Context, _ PubSubMessage) error {
	project := os.Getenv("PROJECT_ID")

	c, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
	if err != nil {
		return err
	}

	computeService, err := compute.New(c)
	if err != nil {
		return err
	}

	req := computeService.Instances.AggregatedList(project)
	if err := req.Pages(ctx, func(page *compute.InstanceAggregatedList) error {
		for _, zoneItems := range page.Items {
			for _, instance := range zoneItems.Instances {
				zoneURL := strings.Split(instance.Zone, "/")
				zone := zoneURL[len(zoneURL)-1]

				log.Printf("Deleting %s in zone %s for project %s\n", instance.Name, zone, project)
				deleteCall := computeService.Instances.Delete(project, zone, instance.Name)
				_, err := deleteCall.Do()
				if err != nil {
					return err
				}
				log.Printf( "Deleted %s in zone %s for project %s\n", instance.Name, zone, project)
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}
