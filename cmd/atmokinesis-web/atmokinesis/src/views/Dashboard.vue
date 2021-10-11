<template>
  <div class="about">
    <div class="container">
      <dashboardTaskTable/>
    </div>
  </div>
</template>

<script>

import DashboardTaskTable from "@/components/DashboardTaskTable";

export default {
  name: 'Dashboard',
  components: {
    'dashboardTaskTable': DashboardTaskTable
  }
}
</script>
