<template>
  <li class="photo-list-entry">
    <img v-bind:src="'//' + host + '/photo/' + photo.filename" />
  </li>
</template>

<script>
export default {
  name: 'photo-list-entry',
  props: ['photo'],
  data () {
    return {
      host: location.host
    }
  }
}
</script>

<style>
  .photo-list-entry > img {
    width: 75%;
    height: auto;
  }
</style>
