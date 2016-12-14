<template>
  <div id="app">
    <photo-list v-bind:photos="photos"></photo-list>
  </div>
</template>

<script>
  import PhotoList from './components/PhotoList'

  export default {
    name: 'app',
    components: {
      PhotoList
    },
    data () {
      return {
        photos: []
      }
    },
    created () {
      this.openWebSocketConnection()
    },
    methods: {
      openWebSocketConnection () {
        let ws = new WebSocket('ws://localhost:6740/data')

        ws.onmessage = (event) => {
          let photo = JSON.parse(event.data)
          if (!photo) return

          this.photos.unshift(photo)
        }
      }
    }

  }

</script>

<style>
    #app {
        font-family: 'Avenir', Helvetica, Arial, sans-serif;
        -webkit-font-smoothing: antialiased;
        -moz-osx-font-smoothing: grayscale;
        text-align: center;
        color: #2c3e50;
        margin-top: 60px;
    }

</style>
