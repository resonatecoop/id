import validateFormdata from '../lib/validateFormdata'
import isEmail from 'validator/lib/isEmail'
import isEmpty from 'validator/lib/isEmpty'
import isLength from 'validator/lib/isLength'
import zxcvbnAsync from 'zxcvbn-async'

// the base form component
export const form = () => ({
  loading: false,
  response: null,
  action: '',
  method: 'POST',
  validator: validateFormdata(),
  valid: true,
  errors: [],
  error: '',

  validate (event) {
    this.validator.validate(event.target.name, event.target.value)
  },

  async submit (event) {
    this.response = null
    this.errors = []

    const formData = new FormData(event.target)

    formData.forEach((value, key) => {
      const shouldValidate = typeof this.validator.validators[key] === 'function'
      if (shouldValidate) {
        this.validator.validate(key, value)
      }
    })

    if (!this.validator.state.valid) {
      // show errors
      for (const [key, value] of Object.entries(this.validator.state.errors)) {
        if (value) {
          this.errors.push({ name: key, message: value.message })
        }
      }
      return
    }

    this.loading = true

    this.errors = []

    const values = this.validator.state.values

    try {
      const response = await fetch('', {
        method: 'POST',
        credentials: 'include',
        headers: {
          Accept: 'application/json',
          'X-CSRF-Token': values['gorilla.csrf.Token'],
          Pragma: 'no-cache',
          'Cache-Control': 'no-cache'
        },
        body: new URLSearchParams(values)
      })

      const isRedirected = response.redirected

      if (isRedirected) {
        window.location.href = response.url
      }

      const status = response.status
      const contentType = response.headers.get('content-type')

      const { error } = await response.json()

      this.error = error

      console.log(status)
      console.log(contentType)
    } catch (err) {
      console.log(err)
    } finally {
      this.loading = false
    }

    console.log(formData)

    this.loading = false
  }
})

export const loginForm = () => Object.assign({}, form(), {
  init () {
    this.validator.field('email', data => {
      if (isEmpty(data)) return new Error('Email is required')
      if (!(isEmail(data))) return new Error('This is not valid email address')
    })

    this.validator.field('password', data => {
      if (isEmpty(data)) return new Error('Password is required')
    })

    this.validator.field('gorilla.csrf.Token', data => {
      if (isEmpty(data)) return new Error('Csrf missing')
    })
  }
})

export const signupForm = () => Object.assign({}, form(), {
  init () {
    this.validator = validateFormdata()

    const zxcvbn = zxcvbnAsync.load({
      sync: true,
      libUrl: 'https://cdn.jsdelivr.net/npm/zxcvbn@4.4.2/dist/zxcvbn.js',
      libIntegrity: 'sha256-9CxlH0BQastrZiSQ8zjdR6WVHTMSA5xKuP5QkEhPNRo='
    })

    this.validator.field('email', (data) => {
      if (isEmpty(data)) {
        return new Error('Please tell us your email address')
      }
      if (!isEmail(data)) {
        return new Error('This is not a valid email address')
      }
    })
    this.validator.field('password', (data) => {
      if (isEmpty(data)) {
        return new Error('A strong password is very important')
      }
      if (!isLength(data, { min: 9 })) {
        return new Error('Password length should not be less than 9 characters')
      }
      const { score, feedback } = zxcvbn(data)
      if (score < 3) {
        return new Error(feedback.warning || (feedback.suggestions.length ? feedback.suggestions[0] : 'Password is too weak'))
      }
      if (!isLength(data, { max: 72 })) {
        return new Error('Password length should not be more than 72 characters')
      }
    })

    this.validator.field('gorilla.csrf.Token', data => {
      if (isEmpty(data)) return new Error('Csrf missing')
    })
  }
})

export const emailUpdateForm = () => Object.assign({}, form(), {
  init () {
    this.validator.field('email', data => {
      if (isEmpty(data)) return new Error('Email is required')
      if (!(isEmail(data))) return new Error('This is not valid email address')
    })

    this.validator.field('password', data => {
      if (isEmpty(data)) return new Error('Password is required')
    })

    this.validator.field('gorilla.csrf.Token', data => {
      if (isEmpty(data)) return new Error('Csrf missing')
    })
  }
})

export const passwordUpdateForm = () => Object.assign({}, form(), {
  init () {
    const zxcvbn = zxcvbnAsync.load({
      sync: true,
      libUrl: 'https://cdn.jsdelivr.net/npm/zxcvbn@4.4.2/dist/zxcvbn.js',
      libIntegrity: 'sha256-9CxlH0BQastrZiSQ8zjdR6WVHTMSA5xKuP5QkEhPNRo='
    })

    this.validator.field('password', { required: !!this.local.token }, (data) => {
      if (isEmpty(data)) return new Error('Current password is required')
      if (/[À-ÖØ-öø-ÿ]/.test(data)) return new Error('Current password may contain unsupported characters. You should ask for a password reset.')
    })
    this.validator.field('password_new', (data) => {
      if (isEmpty(data)) return new Error('New password is required')
      if (data === this.local.data.password) return new Error('Current password and new password are identical')
      const { score, feedback } = zxcvbn(data)
      if (score < 3) {
        return new Error(feedback.warning || (feedback.suggestions.length ? feedback.suggestions[0] : 'Password is too weak'))
      }
      if (!isLength(data, { max: 72 })) {
        return new Error('Password length should not be more than 72 characters')
      }
    })
    this.validator.field('password_confirm', (data) => {
      if (isEmpty(data)) return new Error('Password confirmation is required')
      if (data !== this.local.data.password_new) return new Error('Password mismatch')
    })

    this.validator.field('gorilla.csrf.Token', data => {
      if (isEmpty(data)) return new Error('Csrf missing')
    })
  }
})
