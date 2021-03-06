package operations

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/peter-svintsitskyi/go-gesclient/client"
	"github.com/peter-svintsitskyi/go-gesclient/messages"
	"github.com/peter-svintsitskyi/go-gesclient/tasks"
)

type readStreamEventsForward struct {
	*baseOperation
	stream         string
	start          int
	max            int
	resolveLinkTos bool
	requireMaster  bool
}

func NewReadStreamEventsForward(
	source *tasks.CompletionSource,
	stream string,
	start int,
	max int,
	resolveLinkTos bool,
	requireMaster bool,
	userCredentials *client.UserCredentials,
) *readStreamEventsForward {
	obj := &readStreamEventsForward{
		stream:         stream,
		start:          start,
		max:            max,
		resolveLinkTos: resolveLinkTos,
		requireMaster:  requireMaster,
	}
	obj.baseOperation = newBaseOperation(client.Command_ReadStreamEventsForward,
		client.Command_ReadStreamEventsForwardCompleted, userCredentials, source, obj.createRequestDto,
		obj.inspectResponse, obj.transformResponse, obj.createResponse)
	return obj
}

func (o *readStreamEventsForward) createRequestDto() proto.Message {
	start := int32(o.start)
	max := int32(o.max)
	return &messages.ReadStreamEvents{
		EventStreamId:   &o.stream,
		FromEventNumber: &start,
		MaxCount:        &max,
		ResolveLinkTos:  &o.resolveLinkTos,
		RequireMaster:   &o.requireMaster,
	}
}

func (o *readStreamEventsForward) inspectResponse(message proto.Message) (res *client.InspectionResult, err error) {
	msg := message.(*messages.ReadStreamEventsCompleted)
	switch msg.GetResult() {
	case messages.ReadStreamEventsCompleted_Success,
		messages.ReadStreamEventsCompleted_StreamDeleted,
		messages.ReadStreamEventsCompleted_NoStream:
		err = o.succeed()
	case messages.ReadStreamEventsCompleted_Error:
		err = o.Fail(client.NewServerError(msg.GetError()))
	case messages.ReadStreamEventsCompleted_NotModified:
		err = o.Fail(client.NewNotModified(o.stream))
	case messages.ReadStreamEventsCompleted_AccessDenied:
		err = o.Fail(client.AccessDenied)
	default:
		err = fmt.Errorf("Unexpected ReadStreamResult: %v", *msg.Result)
	}
	if res == nil && err == nil {
		res = client.NewInspectionResult(client.InspectionDecision_EndOperation, msg.GetResult().String(), nil, nil)
	}
	return
}

func (o *readStreamEventsForward) transformResponse(message proto.Message) (interface{}, error) {
	msg := message.(*messages.ReadStreamEventsCompleted)
	status, err := convertStatusCode(msg.GetResult())
	if err != nil {
		return nil, err
	}
	return client.NewStreamEventsSlice(status, o.stream, o.start, client.ReadDirection_Forward, msg.GetEvents(),
		int(msg.GetNextEventNumber()), int(msg.GetLastEventNumber()), msg.GetIsEndOfStream()), nil
}

func (o *readStreamEventsForward) createResponse() proto.Message {
	return &messages.ReadStreamEventsCompleted{}
}

func (o *readStreamEventsForward) String() string {
	return fmt.Sprintf("ReadStreamEventsForward from stream '%s'", o.stream)
}
