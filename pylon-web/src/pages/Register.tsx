import type { ChangeEvent, FormEvent, FocusEvent } from 'react'
import { useState } from 'react'
import { Code, ConnectError } from '@connectrpc/connect'
import { Link, useNavigate } from 'react-router-dom'

import { AuthService } from '../api/auth'
import { useDocumentTitle } from '../hooks/useDocumentTitle'
import { cleanBackendMessage } from '../utils/backendError'

type RegisterFormValues = {
  username: string
  email: string
  password: string
  confirmPassword: string
}

type FieldName = keyof RegisterFormValues
type FieldErrors = Partial<Record<FieldName, string>>
type TouchedFields = Partial<Record<FieldName, boolean>>

const initialValues: RegisterFormValues = {
  username: '',
  email: '',
  password: '',
  confirmPassword: '',
}

const fieldNames: FieldName[] = ['username', 'email', 'password', 'confirmPassword']
const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

export function RegisterPage() {
  useDocumentTitle('Register / Pylon Web')

  const navigate = useNavigate()
  const [values, setValues] = useState<RegisterFormValues>(initialValues)
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

    if (touched[field]) {
      setFieldErrors((current) =>
        updateFieldError(current, field, validateField(field, nextValues)),
      )
    }

    if (field === 'password' && touched.confirmPassword) {
      setFieldErrors((current) =>
        updateFieldError(current, 'confirmPassword', validateField('confirmPassword', nextValues)),
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
      await AuthService.register({
        username: values.username.trim(),
        email: values.email.trim(),
        password: values.password,
      })

      navigate('/login', {
        replace: true,
        state: {
          registrationSuccess: 'Account created. Please log in.',
        },
      })
    } catch (error) {
      const mappedError = mapRegisterError(error)
      setFieldErrors((current) => ({
        ...current,
        ...mappedError.fieldErrors,
      }))
      setGlobalError(mappedError.globalError)
    } finally {
      setIsSubmitting(false)
    }
  }

  function renderField(
    field: FieldName,
    label: string,
    type: 'text' | 'email' | 'password',
    autoComplete: string,
  ) {
    const error = fieldErrors[field]
    const inputId = `register-${field}`
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
            Pylon Web / account creation
          </p>
        </header>

        <section className="col-span-12 grid grid-cols-12 gap-x-4 pt-12 lg:pt-20">
          <div className="col-span-12 lg:col-span-6">
            <p className="mb-5 inline-flex border border-[var(--color-ink)] px-3 py-1 font-mono text-xs uppercase tracking-[0.26em]">
              New operator
            </p>

            <h1 className="max-w-4xl text-[clamp(4rem,14vw,10rem)] font-black uppercase leading-[0.84] tracking-[-0.08em]">
              Register
            </h1>

            <p className="mt-8 max-w-2xl border-t border-[var(--color-line)] pt-5 text-xl font-semibold leading-tight tracking-[-0.03em]">
              Create a Pylon account through the generated Connect auth client.
            </p>
          </div>

          <div className="col-span-12 mt-10 border-y-2 border-[var(--color-ink)] py-6 lg:col-span-6 lg:mt-0">
            {globalError ? (
              <div
                className="mb-6 border-2 border-[var(--color-accent)] px-4 py-3 font-mono text-xs uppercase tracking-[0.16em] text-[var(--color-accent)]"
                role="alert"
              >
                {globalError}
              </div>
            ) : null}

            <form className="space-y-5" noValidate onSubmit={handleSubmit}>
              {renderField('username', 'Username', 'text', 'username')}
              {renderField('email', 'Email', 'email', 'email')}
              {renderField('password', 'Password', 'password', 'new-password')}
              {renderField('confirmPassword', 'Confirm password', 'password', 'new-password')}

              <button
                className="inline-flex w-full items-center justify-between border-2 border-[var(--color-ink)] bg-[var(--color-ink)] px-5 py-4 font-mono text-xs uppercase tracking-[0.22em] text-[var(--color-paper)] transition-transform duration-200 hover:-translate-y-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)] disabled:cursor-not-allowed disabled:opacity-60 disabled:hover:translate-y-0"
                disabled={isSubmitting}
                type="submit"
              >
                {isSubmitting ? (
                  <>
                    Creating account
                    <span
                      aria-hidden="true"
                      className="ml-8 size-4 animate-spin border-2 border-[var(--color-paper)] border-t-transparent"
                    />
                  </>
                ) : (
                  <>
                    Create account
                    <span className="ml-8">↗</span>
                  </>
                )}
              </button>

              <Link
                className="inline-flex w-full items-center justify-between border-2 border-[var(--color-ink)] px-5 py-4 font-mono text-xs uppercase tracking-[0.22em] transition-transform duration-200 hover:-translate-y-1 hover:bg-[var(--color-accent)] hover:text-[var(--color-paper)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
                to="/login"
              >
                Back to login
                <span>→</span>
              </Link>
            </form>
          </div>
        </section>
      </section>
    </main>
  )
}

function validateForm(values: RegisterFormValues): FieldErrors {
  return fieldNames.reduce<FieldErrors>((errors, field) => {
    const error = validateField(field, values)
    if (error) {
      errors[field] = error
    }

    return errors
  }, {})
}

function validateField(field: FieldName, values: RegisterFormValues) {
  switch (field) {
    case 'username':
      if (values.username.trim() === '') {
        return 'Username is required'
      }
      return undefined

    case 'email':
      if (values.email.trim() === '') {
        return 'Email is required'
      }
      if (!emailPattern.test(values.email.trim())) {
        return 'Enter a valid email address'
      }
      return undefined

    case 'password':
      if (values.password === '') {
        return 'Password is required'
      }
      if (values.password.length < 8) {
        return 'Password must be at least 8 characters'
      }
      return undefined

    case 'confirmPassword':
      if (values.confirmPassword === '') {
        return 'Confirm password is required'
      }
      if (values.confirmPassword !== values.password) {
        return "Passwords don't match"
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

function mapRegisterError(error: unknown): {
  fieldErrors: FieldErrors
  globalError: string | null
} {
  const connectError = ConnectError.from(error)
  const message = cleanBackendMessage(connectError.rawMessage || connectError.message)
  const parsedFieldErrors = parseBackendFieldErrors(message)

  if (parsedFieldErrors) {
    return { fieldErrors: parsedFieldErrors, globalError: null }
  }

  if (connectError.code === Code.AlreadyExists) {
    return {
      fieldErrors: {
        email: 'Email or username is already registered',
      },
      globalError: null,
    }
  }

  if (connectError.code === Code.InvalidArgument) {
    const fieldErrors = mapValidationMessageToField(message)
    if (Object.keys(fieldErrors).length > 0) {
      return { fieldErrors, globalError: null }
    }

    return { fieldErrors: {}, globalError: message || 'Registration input is invalid' }
  }

  if (connectError.code === Code.Unavailable) {
    return {
      fieldErrors: {},
      globalError: 'Registration service is unavailable. Please try again.',
    }
  }

  return {
    fieldErrors: {},
    globalError: 'Registration failed. Please try again.',
  }
}

function parseBackendFieldErrors(message: string): FieldErrors | null {
  try {
    const parsed: unknown = JSON.parse(message)
    if (!Array.isArray(parsed)) {
      return null
    }

    const errors: FieldErrors = {}

    for (const item of parsed) {
      if (!isRecord(item)) {
        continue
      }

      const field = typeof item.field === 'string' ? normalizeBackendField(item.field) : null
      const fieldMessage = typeof item.message === 'string' ? item.message.trim() : ''

      if (field && fieldMessage) {
        errors[field] = fieldMessage
      }
    }

    return Object.keys(errors).length > 0 ? errors : null
  } catch {
    return null
  }
}

function mapValidationMessageToField(message: string): FieldErrors {
  const normalizedMessage = message.toLowerCase()

  if (normalizedMessage.includes('username')) {
    return { username: message }
  }

  if (normalizedMessage.includes('email')) {
    return { email: message }
  }

  if (normalizedMessage.includes('password')) {
    return { password: message }
  }

  return {}
}

function normalizeBackendField(field: string): FieldName | null {
  const normalizedField = field.trim().toLowerCase().replace(/[-_]/g, '')

  switch (normalizedField) {
    case 'username':
      return 'username'
    case 'email':
      return 'email'
    case 'password':
      return 'password'
    case 'confirmpassword':
      return 'confirmPassword'
    default:
      return null
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}
