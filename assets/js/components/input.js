export const input = () => ({
  error: null,

  validate () {
    if (!this.response?.errors?.[this.$el.name]) return (this.error = null)
    this.error = this.response.errors[this.$el.name]
  }
})
