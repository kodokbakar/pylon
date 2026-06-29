import type { PartialMessage } from '@bufbuild/protobuf'
import { createPromiseClient } from '@connectrpc/connect'

import {
  CreateRoomRequest,
  type CreateRoomResponse,
  GetRoomMembersRequest,
  type GetRoomMembersResponse,
  GetRoomRequest,
  type GetRoomResponse,
  LeaveRoomRequest,
  ListRoomsRequest,
  type ListRoomsResponse,
} from './gen/pylon/room/v1/room_service_pb'
import { RoomService as RoomServiceDefinition } from './gen/pylon/room/v1/room_service_connect'
import { authenticatedTransport } from './transport'

const client = createPromiseClient(RoomServiceDefinition, authenticatedTransport)

async function createRoom(input: PartialMessage<CreateRoomRequest>): Promise<CreateRoomResponse> {
  return client.createRoom(input)
}

async function listRooms(input: PartialMessage<ListRoomsRequest>): Promise<ListRoomsResponse> {
  return client.listRooms(input)
}

async function getRoom(input: PartialMessage<GetRoomRequest>): Promise<GetRoomResponse> {
  return client.getRoom(input)
}

async function getRoomMembers(
  input: PartialMessage<GetRoomMembersRequest>,
): Promise<GetRoomMembersResponse> {
  return client.getRoomMembers(input)
}

async function leaveRoom(input: PartialMessage<LeaveRoomRequest>): Promise<void> {
  await client.leaveRoom(input)
}

export const RoomService = {
  createRoom,
  listRooms,
  getRoom,
  getRoomMembers,
  leaveRoom,
}
