<template>
  <div style="padding: 24px">
    <h2 style="margin: 0 0 20px">岗位管理</h2>

    <el-form inline style="margin-bottom: 16px">
      <el-form-item label="类型">
        <el-select v-model="filter.biz_type" style="width: 90px">
          <el-option label="全部" :value="0" />
          <el-option label="招聘" :value="1" />
          <el-option label="求职" :value="2" />
        </el-select>
      </el-form-item>
      <el-form-item label="状态">
        <el-select v-model="filter.status" multiple collapse-tags style="width: 160px" placeholder="全部">
          <el-option label="活跃" :value="1" />
          <el-option label="已关闭" :value="2" />
          <el-option label="已禁用" :value="3" />
          <el-option label="已删除" :value="4" />
        </el-select>
      </el-form-item>
      <el-form-item label="关键词">
        <el-input v-model="filter.keyword" placeholder="职位 / 公司名" style="width: 180px" clearable @keyup.enter="load(1)" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="load(1)">搜索</el-button>
      </el-form-item>
    </el-form>

    <el-table :data="list" v-loading="loading" border stripe row-key="job_id">
      <el-table-column prop="job_id" label="ID" width="90" />
      <el-table-column label="类型" width="70">
        <template #default="{ row }">
          <el-tag :type="row.biz_type === 1 ? '' : 'warning'" size="small">{{ row.biz_type === 1 ? '招聘' : '求职' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="positions" label="职位" width="120" show-overflow-tooltip />
      <el-table-column prop="description" label="岗位描述" min-width="180" show-overflow-tooltip />
      <el-table-column label="图片" width="140">
        <template #default="{ row }">
          <div v-if="row.photo_urls?.length" class="photo-stack">
            <el-image
              v-for="(url, i) in row.photo_urls.slice(0, 3)"
              :key="i"
              :src="url"
              :preview-src-list="row.photo_urls"
              :initial-index="i"
              preview-teleported
              class="photo-stack-item"
              :style="{ left: i * 28 + 'px', zIndex: 3 - i }"
              fit="cover"
            />
            <span
              v-if="row.photo_urls.length > 3"
              class="photo-stack-more"
              :style="{ left: 3 * 28 + 'px' }"
            >+{{ row.photo_urls.length - 3 }}</span>
          </div>
          <span v-else style="color:#ccc">—</span>
        </template>
      </el-table-column>
      <el-table-column prop="company_name" label="公司" width="140" show-overflow-tooltip />
      <el-table-column label="地区" width="160" show-overflow-tooltip>
        <template #default="{ row }">{{ [row.first_area_des, row.second_area_des, row.third_area_des].filter(Boolean).join(' ') }}</template>
      </el-table-column>
      <el-table-column label="薪资(元/月)" width="120">
        <template #default="{ row }">{{ row.salary_min }}-{{ row.salary_max }}</template>
      </el-table-column>
      <el-table-column label="状态" width="85">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="user_name" label="发布人" width="100" />
      <el-table-column prop="user_phone" label="手机号" width="120" />
      <el-table-column prop="create_at" label="发布时间" width="160" />
      <el-table-column prop="update_at" label="更新时间" width="160" />
      <el-table-column label="操作" fixed="right" width="190">
        <template #default="{ row }">
          <el-button v-if="row.status !== 3" size="small" type="warning" @click="doAction('disable', row)">禁用</el-button>
          <el-button v-else size="small" type="success" @click="doAction('enable', row)">恢复</el-button>
          <el-popconfirm title="确认删除该岗位？" @confirm="doAction('delete', row)">
            <template #reference>
              <el-button size="small" type="danger">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      style="margin-top: 16px; display: flex; justify-content: flex-end"
      layout="total, sizes, prev, pager, next"
      :total="total"
      v-model:current-page="page"
      v-model:page-size="pageSize"
      :page-sizes="[20, 50, 100]"
      @change="load"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { listJobs, disableJob, enableJob, deleteJob } from './api'

const list = ref([])
const total = ref(0)
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const filter = ref({ biz_type: 0, status: [], keyword: '' })

const statusLabel = (s) => ({ 1: '活跃', 2: '已关闭', 3: '已禁用', 4: '已删除' }[s] ?? s)
const statusType = (s) => ({ 1: 'success', 2: 'info', 3: 'danger', 4: 'warning' }[s])

async function load(p) {
  if (typeof p === 'number') page.value = p
  loading.value = true
  try {
    const res = await listJobs({ ...filter.value, page_num: page.value, page_size: pageSize.value })
    list.value = res.data?.list ?? []
    total.value = res.data?.total ?? 0
  } finally {
    loading.value = false
  }
}

async function doAction(type, row) {
  const fn = { disable: disableJob, enable: enableJob, delete: deleteJob }[type]
  await fn(row.job_id)
  ElMessage.success('操作成功')
  load()
}

onMounted(load)
</script>

<style scoped>
.photo-stack {
  position: relative;
  display: flex;
  align-items: center;
  height: 44px;
}

.photo-stack-item {
  position: absolute;
  width: 40px;
  height: 40px;
  border-radius: 6px;
  border: 2px solid #fff;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.12);
  cursor: pointer;
  transition: transform 0.2s, z-index 0.2s;
}

.photo-stack-item:hover {
  transform: scale(1.1);
  z-index: 10 !important;
}

.photo-stack-more {
  position: absolute;
  width: 40px;
  height: 40px;
  border-radius: 6px;
  background: rgba(0, 0, 0, 0.45);
  color: #fff;
  font-size: 12px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid #fff;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.12);
}
</style>
