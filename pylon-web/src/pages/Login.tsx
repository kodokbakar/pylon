import type { ChangeEvent, FormEvent, FocusEvent } from 'react'
import { useState } from 'react'
import { Code, ConnectError } from '@connectrpc/connect'
import { Link, useLocation, useNavigate } from 'react-router-dom'

import { AuthService } from '../api/auth'
import type { User } from '../api/gen/pylon/auth/v1/auth_service_pb'
import { useAuth } from '../hooks/useAuth'
import { useDocumentTitle } from '../hooks/useDocumentTitle'
import { cleanBackendMessage } from '../utils/backendError'
import type { StoredAuthUser } from '../utils/authToken'

type LoginFormValues = {
  email: string
  password: string
}

type FieldName = keyof LoginFormValues
type FieldErrors = Partial<Record<FieldName, string>>
type TouchedFields = Partial<Record<FieldName, boolean>>

type LoginLocationState = {
  registrationSuccess?: string
}

const initialValues: LoginFormValues = {
  email: '',
  password: '',
}

const fieldNames: FieldName[] = ['email', 'password']

export function LoginPage() {
  useDocumentTitle('Login / Pylon Web')

  const navigate = useNavigate()
  const location = useLocation()
  const { login } = useAuth()
  const locationState = location.state as LoginLocationState | null
  const registrationSuccess = locationState?.registrationSuccess

  const [values, setValues] = useState<LoginFormValues>(initialValues)
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [touched, setTouched] = useState<TouchedFields>({})
  const [globalError, setGlobalError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  function handleChange(event: ChangeEvent<HTMLInputElement>) {
    const field = event.target.name as FieldName
    const nextValues = {
      ...values,
      [field]: event.target.value,
    }

    setValues(nextValues)
    setGlobalError(null)

    if (touched[field]) {
      setFieldErrors((current) =>
        updateFieldError(current, field, validateField(field, nextValues)),
      )
    }
  }

  function handleBlur(event: FocusEvent<HTMLInputElement>) {
    const field = event.target.name as FieldName

    setTouched((current) => ({
      ...current,
      [field]: true,
    }))

    setFieldErrors((current) => updateFieldError(current, field, validateField(field, values)))
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const nextErrors = validateForm(values)
    setTouched(
      fieldNames.reduce<TouchedFields>((result, field) => ({ ...result, [field]: true }), {}),
    )
    setFieldErrors(nextErrors)
    setGlobalError(null)

    if (Object.keys(nextErrors).length > 0) {
      return
    }

    setIsSubmitting(true)

    try {
      const response = await AuthService.login({
        email: values.email.trim(),
        password: values.password,
      })

      const token = response.token.trim()
      const refreshToken = response.refreshToken.trim()

      if (!token || !refreshToken) {
        setGlobalError('Login response did not include a complete auth session.')
        return
      }

      login({
        token,
        refreshToken,
        user: toStoredAuthUser(response.user),
      })

      navigate('/', { replace: true })
    } catch (error) {
      setGlobalError(mapLoginError(error))
    } finally {
      setIsSubmitting(false)
    }
  }

  function renderField(
    field: FieldName,
    label: string,
    type: 'email' | 'password',
    autoComplete: string,
  ) {
    const error = fieldErrors[field]
    const inputId = `login-${field}`
    const errorId = `${inputId}-error`

    return (
      <div>
        <label
          className="font-mono text-xs uppercase tracking-[0.22em] text-[var(--color-muted)]"
          htmlFor={inputId}
        >
          {label}
        </label>

        <input
          aria-describedby={error ? errorId : undefined}
          aria-invalid={Boolean(error)}
          autoComplete={autoComplete}
          className="mt-3 w-full border-2 border-[var(--color-ink)] bg-transparent px-4 py-4 text-base font-semibold outline-none transition-colors placeholder:text-[var(--color-muted)] focus:bg-[var(--color-grid)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)] disabled:cursor-not-allowed disabled:opacity-60"
          disabled={isSubmitting}
          id={inputId}
          name={field}
          placeholder={label}
          type={type}
          value={values[field]}
          onBlur={handleBlur}
          onChange={handleChange}
        />

        {error ? (
          <p
            className="mt-2 font-mono text-xs uppercase tracking-[0.16em] text-[var(--color-accent)]"
            id={errorId}
          >
            {error}
          </p>
        ) : null}
      </div>
    )
  }

  return (
    <main className="min-h-screen bg-[var(--color-paper)] text-[var(--color-ink)]">
      <section className="mx-auto grid min-h-screen w-full max-w-7xl grid-cols-12 gap-x-4 px-5 py-6 sm:px-8 lg:px-10">
        <header className="col-span-12 border-b border-[var(--color-line)] pb-4">
          <p className="font-mono text-xs uppercase tracking-[0.34em] text-[var(--color-muted)]">
            Pylon Web / authentication
          </p>
        </header>

        <section className="col-span-12 grid grid-cols-12 gap-x-4 pt-12 lg:pt-20">
          <div className="col-span-12 lg:col-span-6">
            <p className="mb-5 inline-flex border border-[var(--color-ink)] px-3 py-1 font-mono text-xs uppercase tracking-[0.26em]">
              Access terminal
            </p>

            <h1 className="max-w-4xl text-[clamp(4rem,14vw,10rem)] font-black uppercase leading-[0.84] tracking-[-0.08em]">
              Login
            </h1>

            <p className="mt-8 max-w-2xl border-t border-[var(--color-line)] pt-5 text-xl font-semibold leading-tight tracking-[-0.03em]">
              Authenticate through the generated Connect auth client and enter the protected app.
            </p>
          </div>

          <div className="col-span-12 mt-10 border-y-2 border-[var(--color-ink)] py-6 lg:col-span-6 lg:mt-0">
            {registrationSuccess ? (
              <div
                className="mb-6 border-2 border-[var(--color-accent)] px-4 py-3 font-mono text-xs uppercase tracking-[0.16em] text-[var(--color-accent)]"
                role="status"
              >
                {registrationSuccess}
              </div>
            ) : null}

            {globalError ? (
              <div
                className="mb-6 border-2 border-[var(--color-accent)] px-4 py-3 font-mono text-xs uppercase tracking-[0.16em] text-[var(--color-accent)]"
                role="alert"
              >
                {globalError}
              </div>
            ) : null}

            <form className="space-y-5" noValidate onSubmit={handleSubmit}>
              {renderField('email', 'Email', 'email', 'email')}
              {renderField('password', 'Password', 'password', 'current-password')}

              <button
                className="inline-flex w-full items-center justify-between border-2 border-[var(--color-ink)] bg-[var(--color-ink)] px-5 py-4 font-mono text-xs uppercase tracking-[0.22em] text-[var(--color-paper)] transition-transform duration-200 hover:-translate-y-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)] disabled:cursor-not-allowed disabled:opacity-60 disabled:hover:translate-y-0"
                disabled={isSubmitting}
                type="submit"
              >
                {isSubmitting ? (
                  <>
                    Signing in
                    <span
                      aria-hidden="true"
                      className="ml-8 size-4 animate-spin border-2 border-[var(--color-paper)] border-t-transparent"
                    />
                  </>
                ) : (
                  <>
                    Sign in
                    <span className="ml-8">↗</span>
                  </>
                )}
              </button>

              <Link
                className="inline-flex w-full items-center justify-between border-2 border-[var(--color-ink)] px-5 py-4 font-mono text-xs uppercase tracking-[0.22em] transition-transform duration-200 hover:-translate-y-1 hover:bg-[var(--color-accent)] hover:text-[var(--color-paper)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
                to="/register"
              >
                Don&apos;t have an account? Register
                <span>→</span>
              </Link>
            </form>
          </div>
        </section>
      </section>
    </main>
  )
}

function validateForm(values: LoginFormValues): FieldErrors {
  return fieldNames.reduce<FieldErrors>((errors, field) => {
    const error = validateField(field, values)
    if (error) {
      errors[field] = error
    }

    return errors
  }, {})
}

function validateField(field: FieldName, values: LoginFormValues) {
  switch (field) {
    case 'email':
      if (values.email.trim() === '') {
        return 'Email is required'
      }
      return undefined

    case 'password':
      if (values.password === '') {
        return 'Password is required'
      }
      return undefined
  }
}

function updateFieldError(errors: FieldErrors, field: FieldName, error: string | undefined) {
  const nextErrors = { ...errors }

  if (error) {
    nextErrors[field] = error
  } else {
    delete nextErrors[field]
  }

  return nextErrors
}

function mapLoginError(error: unknown) {
  const connectError = ConnectError.from(error)

  if (connectError.code === Code.Unauthenticated) {
    return 'Invalid email or password.'
  }

  if (connectError.code === Code.InvalidArgument) {
    return (
      cleanBackendMessage(connectError.rawMessage || connectError.message) ||
      'Login input is invalid.'
    )
  }

  if (connectError.code === Code.Unavailable) {
    return 'Login service is unavailable. Please try again.'
  }

  return 'Login failed. Please try again.'
}

function toStoredAuthUser(user: User | undefined): StoredAuthUser | null {
  if (!user) {
    return null
  }

  return {
    id: user.id,
    username: user.username,
    email: user.email,
    displayName: user.displayName,
    avatarUrl: user.avatarUrl,
    createdAt: user.createdAt,
  }
}
