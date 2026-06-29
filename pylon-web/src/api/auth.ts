import type { PartialMessage } from '@bufbuild/protobuf'
import { createPromiseClient } from '@connectrpc/connect'

import {
  LoginRequest,
  type LoginResponse,
  RefreshTokenRequest,
  type RefreshTokenResponse,
  RegisterRequest,
  type RegisterResponse,
} from './gen/pylon/auth/v1/auth_service_pb'
import { AuthService as AuthServiceDefinition } from './gen/pylon/auth/v1/auth_service_connect'
import { publicTransport } from './transport'

const client = createPromiseClient(AuthServiceDefinition, publicTransport)

async function register(input: PartialMessage<RegisterRequest>): Promise<RegisterResponse> {
  return client.register(input)
}

async function login(input: PartialMessage<LoginRequest>): Promise<LoginResponse> {
  return client.login(input)
}

async function refreshToken(
  input: PartialMessage<RefreshTokenRequest>,
): Promise<RefreshTokenResponse> {
  return client.refreshToken(input)
}

export const AuthService = {
  register,
  login,
  refreshToken,
}
