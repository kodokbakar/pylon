import type { PartialMessage } from '@bufbuild/protobuf'
import { createPromiseClient } from '@connectrpc/connect'

import {
  GetRoomRequest,
  type GetRoomResponse,
  ListRoomsRequest,
  type ListRoomsResponse,
} from './gen/pylon/room/v1/room_service_pb'
import { RoomService as RoomServiceDefinition } from './gen/pylon/room/v1/room_service_connect'
import { authenticatedTransport } from './transport'

const client = createPromiseClient(RoomServiceDefinition, authenticatedTransport)

async function listRooms(input: PartialMessage<ListRoomsRequest>): Promise<ListRoomsResponse> {
  return client.listRooms(input)
}

async function getRoom(input: PartialMessage<GetRoomRequest>): Promise<GetRoomResponse> {
  return client.getRoom(input)
}

export const RoomService = {
  listRooms,
  getRoom,
}
