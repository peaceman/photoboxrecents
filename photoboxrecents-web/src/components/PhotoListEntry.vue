<template>
  <li>
    <img v-bind:src="'http://localhost:6740/photo/' + photo.filename" />
  </li>
</template>

<script>
export default {
  name: 'photo-list-entry',
  props: ['photo']
}
</script>
