package events

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	v1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

var Log = log.WithField("system", "events")

type EventRouter struct {
	kubeClient     kubernetes.Interface
	eLister        corelisters.EventLister
	eListerSynched cache.InformerSynced
	skins          []metrics.DataSink
}

type EventSinkInterface interface {
	UpdateEvents(function string, eNew *v1.Event, eOld *v1.Event)
}

func CreateEventRouter(clientset kubernetes.Interface, skins []metrics.DataSink, clusterName string) (*EventRouter, informers.SharedInformerFactory) {
	sharedInformers := informers.NewSharedInformerFactory(clientset, time.Minute)
	eventsInformer := sharedInformers.Core().V1().Events()
	eventRouter := newEventRouter(clientset, eventsInformer, skins, clusterName)
	return eventRouter, sharedInformers
}

func newEventRouter(kubeClient kubernetes.Interface, eventsInformer coreinformers.EventInformer, skins []metrics.DataSink, clusterName string) *EventRouter {
	er := &EventRouter{
		kubeClient: kubeClient,
		skins:      skins,
	}
	eventsInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: er.addEvent,
		// UpdateFunc: er.updateEvent,
		// DeleteFunc: er.deleteEvent,
	})
	er.eLister = eventsInformer.Lister()
	er.eListerSynched = eventsInformer.Informer().HasSynced
	return er
}

// Run starts the events/Controller.
func (er *EventRouter) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer log.Infof("Shutting down EventRouter")

	log.Infof("Starting EventRouter")

	// here is where we kick the caches into gear
	if !cache.WaitForCacheSync(stopCh, er.eListerSynched) {
		utilruntime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}
	<-stopCh
}

// addEvent is called when an event is created, or during the initial list
func (er *EventRouter) addEvent(obj interface{}) {
	e := obj.(*v1.Event)
	for _, skin := range er.skins {
		skin.ExportEvents("added", e, nil)
	}
}

// updateEvent is called any time there is an update to an existing event
func (er *EventRouter) updateEvent(objOld interface{}, objNew interface{}) {
	eOld := objOld.(*v1.Event)
	eNew := objNew.(*v1.Event)
	for _, skin := range er.skins {
		skin.ExportEvents("update", eNew, eOld)
	}
}

// deleteEvent should only occur when the system garbage collects events via TTL expiration
func (er *EventRouter) deleteEvent(obj interface{}) {
	e := obj.(*v1.Event)
	for _, skin := range er.skins {
		skin.ExportEvents("delete", e, nil)
	}
}
