export const search = () => ({
  tags: [
    'ambient',
    'acoustic',
    'alternative',
    'chill',
    'dream-pop',
    'electro',
    'electronic',
    'experimental',
    'folk',
    'funk',
    'hiphop',
    'house',
    'indie-pop',
    'indie-rock',
    'instrumental',
    'jazz',
    'metal',
    'podcasts',
    'pop',
    'punk',
    'reggae'
  ],

  query: '',

  open: true,

  init () {
    this.$watch('open', () => {
      if (this.open) {
        this.$nextTick(() => this.$refs.input.focus())
      }
    })
  },

  submit (event) {
    const q = event.target.search.value

    if (!q) return false
    if (q.length < 3) return false

    const bang = q.startsWith('#')
    const pathname = bang ? '/tag' : '/search'
    const url = new URL(pathname, this.$store.app.state.appURL || 'http://localhost')
    const params = bang ? { term: q.split('#')[1] } : { q }
    url.search = new URLSearchParams(params)
    window.open(url.href, '_blank')
    return false
  }
})
