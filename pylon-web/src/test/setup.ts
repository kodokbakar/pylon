import '@testing-library/jest-dom/vitest'
import { afterEach } from 'vitest'

import './mocks/server'

afterEach(() => {
  window.localStorage.clear()
  window.sessionStorage.clear()
})
