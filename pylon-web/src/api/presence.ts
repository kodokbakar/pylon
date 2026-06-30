import type { PartialMessage } from '@bufbuild/protobuf'
import { createPromiseClient, type CallOptions } from '@connectrpc/connect'

import {
  GetRoomPresenceRequest,
  type GetRoomPresenceResponse,
  type PresenceEvent,
  SetTypingRequest,
  StreamPresenceRequest,
} from './gen/pylon/presence/v1/presence_service_pb'
import { PresenceService as PresenceServiceDefinition } from './gen/pylon/presence/v1/presence_service_connect'
import { authenticatedTransport } from './transport'

const client = createPromiseClient(PresenceServiceDefinition, authenticatedTransport)

async function getRoomPresence(
  input: PartialMessage<GetRoomPresenceRequest>,
): Promise<GetRoomPresenceResponse> {
  return client.getRoomPresence(input)
}

async function setTyping(input: PartialMessage<SetTypingRequest>): Promise<void> {
  await client.setTyping(input)
}

function streamPresence(
  input: PartialMessage<StreamPresenceRequest>,
  options?: CallOptions,
): AsyncIterable<PresenceEvent> {
  return client.streamPresence(input, options)
}

export const PresenceService = {
  getRoomPresence,
  setTyping,
  streamPresence,
}
