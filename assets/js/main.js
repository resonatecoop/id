import Alpine from 'alpinejs'
import { loginForm } from './components/form'
import { input } from './components/input'
import { search } from './components/search'

// directives
Alpine.directive('uppercase', el => {
  el.textContent = el.textContent.toUpperCase()
})

// https://alpinejs.dev/advanced/extending#evaluating-expressions
Alpine.directive('log', (el, { expression }, { evaluateLater, effect }) => {
  const getThingToLog = evaluateLater(expression)

  effect(() => {
    getThingToLog(thingToLog => {
      console.log(thingToLog)
    })
  })
})

// Await Alpine.js initialization
document.addEventListener('alpine:init', () => {
  Alpine.store('darkMode', {
    on: false,

    toggle () {
      this.on = !this.on
    },

    init () {
      this.on = window.matchMedia('(prefers-color-scheme: dark)').matches
    }
  })
  Alpine.store('app', {
    init () {
      this.state = Object.assign({}, window.initialState)
      delete window.initialState
    }
  })
  Alpine.data('loginForm', loginForm)
  Alpine.data('search', search)
  Alpine.data('input', input)
})

// dev only?
window.Alpine = Alpine

Alpine.start()
