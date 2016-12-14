<template>
  <div class="photo-list">
    <ul>
      <photo-list-entry v-for="photo in photos" v-bind:photo="photo"></photo-list-entry>
    </ul>
  </div>
</template>

<script>
import PhotoListEntry from './PhotoListEntry'

export default {
  name: 'photo-list',
  components: {
    PhotoListEntry
  },
  props: ['photos']
}
</script>

<style>
  .photo-list > ul {
    list-style: none;
    padding: 0;
  }
</style>
