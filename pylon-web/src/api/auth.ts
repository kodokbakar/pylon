import type { PartialMessage } from '@bufbuild/protobuf'
import { createPromiseClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'

import {
  LoginRequest,
  type LoginResponse,
  RegisterRequest,
  type RegisterResponse,
} from './gen/pylon/auth/v1/auth_service_pb'
import { AuthService as AuthServiceDefinition } from './gen/pylon/auth/v1/auth_service_connect'
import { getApiBaseUrl } from './config'

const transport = createConnectTransport({
  baseUrl: getApiBaseUrl(),
})

const client = createPromiseClient(AuthServiceDefinition, transport)

async function register(input: PartialMessage<RegisterRequest>): Promise<RegisterResponse> {
  return client.register(input)
}

async function login(input: PartialMessage<LoginRequest>): Promise<LoginResponse> {
  return client.login(input)
}

export const AuthService = {
  register,
  login,
}
