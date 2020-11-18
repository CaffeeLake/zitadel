package org

import (
	"context"

	"github.com/caos/zitadel/internal/eventstore/v2"
	"github.com/caos/zitadel/internal/v2/repository/member"
	"github.com/caos/zitadel/internal/v2/repository/members"
)

const (
	orgEventTypePrefix = eventstore.EventType("org.")
)

var (
	MemberAddedEventType   = orgEventTypePrefix + member.AddedEventType
	MemberChangedEventType = orgEventTypePrefix + member.ChangedEventType
	MemberRemovedEventType = orgEventTypePrefix + member.RemovedEventType
)

type MemberWriteModel struct {
	member.WriteModel
}

// func NewMemberAggregate(userID string) *MemberAggregate {
// 	return &MemberAggregate{
// 		Aggregate: member.NewAggregate(
// 			eventstore.NewAggregate(userID, MemberAggregateType, "RO", AggregateVersion, 0),
// 		),
// 		// Aggregate: member.NewMemberAggregate(userID),
// 	}
// }

type MembersReadModel struct {
	members.ReadModel
}

func (rm *MembersReadModel) AppendEvents(events ...eventstore.EventReader) (err error) {
	for _, event := range events {
		switch e := event.(type) {
		case *MemberAddedEvent:
			rm.ReadModel.AppendEvents(&e.AddedEvent)
		case *MemberChangedEvent:
			rm.ReadModel.AppendEvents(&e.ChangedEvent)
		case *MemberRemovedEvent:
			rm.ReadModel.AppendEvents(&e.RemovedEvent)
		}
	}
	return nil
}

type MemberReadModel member.ReadModel

func (rm *MemberReadModel) AppendEvents(events ...eventstore.EventReader) (err error) {
	for _, event := range events {
		switch e := event.(type) {
		case *MemberAddedEvent:
			rm.ReadModel.AppendEvents(&e.AddedEvent)
		case *MemberChangedEvent:
			rm.ReadModel.AppendEvents(&e.ChangedEvent)
		}
	}
	return nil
}

type MemberAddedEvent struct {
	member.AddedEvent
}

type MemberChangedEvent struct {
	member.ChangedEvent
}
type MemberRemovedEvent struct {
	member.RemovedEvent
}

func NewMemberAddedEvent(
	ctx context.Context,
	userID string,
	roles ...string,
) *MemberAddedEvent {

	return &MemberAddedEvent{
		AddedEvent: *member.NewAddedEvent(
			eventstore.NewBaseEventForPush(
				ctx,
				MemberAddedEventType,
			),
			userID,
			roles...,
		),
	}
}

func NewMemberChangedEvent(
	ctx context.Context,
	current,
	changed *MemberWriteModel,
) (*MemberChangedEvent, error) {

	event, err := member.NewChangedEvent(
		eventstore.NewBaseEventForPush(
			ctx,
			MemberChangedEventType,
		),
		&current.WriteModel,
		&changed.WriteModel,
	)
	if err != nil {
		return nil, err
	}
	return &MemberChangedEvent{
		ChangedEvent: *event,
	}, nil
}

func NewMemberRemovedEvent(
	ctx context.Context,
	userID string,
) *MemberRemovedEvent {

	return &MemberRemovedEvent{
		RemovedEvent: *member.NewRemovedEvent(
			eventstore.NewBaseEventForPush(
				ctx,
				MemberRemovedEventType,
			),
			userID,
		),
	}
}
