package events

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/dig"
	"golang.org/x/exp/slices"
)

type IDomainEventHandler interface {
	Handle(ctx context.Context, domainEvent IDomainEvent)
}

type IDomainEvent interface {
}

type IEventHandler interface {
	Handle(context.Context, IEvent)
}

type IEvent interface {
}

type IEventDispatcher interface {
	AddDomainEvent(IDomainEvent)
	CommitDomainEventsStack(ctx context.Context)
	DispatchEvent(context.Context, IEvent)
	AddEvent(IEvent)
	CommitEventsStack(ctx context.Context)
}

type EventDispatcherParams struct {
	dig.In

	DomainEventHandlers []IDomainEventHandler `group:"DomainEventHandlers"`
	EventHandlers       []IEventHandler       `group:"EventHandlers"`
}

type EventDispatcher struct {
	domainEventHandlers []IDomainEventHandler
	eventHandlers       []IEventHandler

	domainEvents []IDomainEvent
	events       []IEvent
}

func (eventDispatcher *EventDispatcher) AddDomainEvent(event IDomainEvent) {
	eventDispatcher.domainEvents = append(eventDispatcher.domainEvents, event)
}

func (eventDispatcher *EventDispatcher) AddEvent(event IEvent) {
	eventDispatcher.events = append(eventDispatcher.events, event)
}

func (eventDispatcher *EventDispatcher) dispatchDomainEvent(ctx context.Context, event IDomainEvent) {

	position := slices.IndexFunc(eventDispatcher.domainEventHandlers, func(handler IDomainEventHandler) bool {
		handlerName := fmt.Sprintf("%T", handler)
		eventName := fmt.Sprintf("%T", event)
		return strings.Contains(handlerName, eventName)
	})

	eventDispatcher.domainEventHandlers[position].Handle(ctx, event)
}

func (eventDispatcher *EventDispatcher) DispatchEvent(ctx context.Context, event IEvent) {
	position := slices.IndexFunc(eventDispatcher.eventHandlers, func(handler IEventHandler) bool {
		handlerName := fmt.Sprintf("%T", handler)
		eventName := fmt.Sprintf("%T", event)
		return strings.Contains(handlerName, eventName)
	})

	eventDispatcher.eventHandlers[position].Handle(ctx, event)
}

func (eventDispatcher *EventDispatcher) CommitDomainEventsStack(ctx context.Context) {
	for _, event := range eventDispatcher.domainEvents {
		eventDispatcher.dispatchDomainEvent(ctx, event)
	}
	eventDispatcher.clearDomainEvents()
}

func (eventDispatcher *EventDispatcher) clearDomainEvents() {
	eventDispatcher.domainEvents = make([]IDomainEvent, 0)
}

func (eventDispatcher *EventDispatcher) CommitEventsStack(ctx context.Context) {
	for _, event := range eventDispatcher.events {
		eventDispatcher.DispatchEvent(ctx, event)
	}

	eventDispatcher.clearEvents()
}

func (eventDispatcher *EventDispatcher) clearEvents() {
	eventDispatcher.events = make([]IEvent, 0)
}

func NewEventDispatcher(params EventDispatcherParams) IEventDispatcher {
	return &EventDispatcher{domainEventHandlers: params.DomainEventHandlers, eventHandlers: params.EventHandlers}
}
