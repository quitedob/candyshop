<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-semibold text-gray-900">{{ t('admin.oem.title') }}</h1>
    </div>

    <!-- Tabs -->
    <div class="border-b border-gray-200 mb-6">
      <nav class="-mb-px flex space-x-8">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          @click="activeTab = tab.key"
          :class="[
            'pb-3 px-1 text-sm font-medium border-b-2 transition-colors',
            activeTab === tab.key
              ? 'border-orange-500 text-orange-600'
              : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
          ]"
        >{{ tab.label }}</button>
      </nav>
    </div>

    <!-- Flows Tab -->
    <div v-if="activeTab === 'flows'">
      <div class="flex justify-between items-center mb-4">
        <p class="text-sm text-gray-500">{{ flows.length }} {{ t('admin.oem.tab_flows') }}</p>
        <button @click="openFlowForm()" class="btn-primary text-sm px-4 py-2 rounded-md bg-orange-500 text-white hover:bg-orange-600">
          {{ t('admin.oem.new_flow') }}
        </button>
      </div>

      <div v-if="flowsLoading" class="text-center py-10 text-gray-500">{{ t('admin.oem.loading') }}</div>
      <div v-else-if="!flows.length" class="text-center py-10 text-gray-500">{{ t('admin.oem.no_flows') }}</div>
      <table v-else class="min-w-full bg-white rounded-lg shadow">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.oem.col_title') }}</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.oem.col_type') }}</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.oem.col_steps') }}</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.oem.col_actions') }}</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200">
          <tr v-for="flow in flows" :key="flow.id">
            <td class="px-4 py-3 text-sm">{{ flow.title }}</td>
            <td class="px-4 py-3 text-sm">
              <span :class="flow.type === 'quick_odm' ? 'bg-blue-100 text-blue-700' : 'bg-purple-100 text-purple-700'" class="px-2 py-0.5 rounded-full text-xs font-medium">
                {{ flow.type === 'quick_odm' ? t('admin.oem.type_quick_odm') : t('admin.oem.type_full_oem') }}
              </span>
            </td>
            <td class="px-4 py-3 text-sm text-gray-500">{{ flow.steps?.length || 0 }}</td>
            <td class="px-4 py-3 text-sm space-x-2">
              <button @click="openFlowForm(flow)" class="text-orange-600 hover:underline text-xs">{{ t('admin.oem.edit') }}</button>
              <button @click="deleteFlow(flow.id)" class="text-red-600 hover:underline text-xs">{{ t('admin.oem.delete') }}</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Solutions Tab -->
    <div v-if="activeTab === 'solutions'">
      <div class="flex justify-between items-center mb-4">
        <p class="text-sm text-gray-500">{{ solutions.length }} {{ t('admin.oem.tab_solutions') }}</p>
        <button @click="openSolutionForm()" class="btn-primary text-sm px-4 py-2 rounded-md bg-orange-500 text-white hover:bg-orange-600">
          {{ t('admin.oem.new_solution') }}
        </button>
      </div>

      <div v-if="solutionsLoading" class="text-center py-10 text-gray-500">{{ t('admin.oem.loading') }}</div>
      <div v-else-if="!solutions.length" class="text-center py-10 text-gray-500">{{ t('admin.oem.no_solutions') }}</div>
      <table v-else class="min-w-full bg-white rounded-lg shadow">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.oem.col_title') }}</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.oem.col_category') }}</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.oem.col_moq') }}</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.oem.col_actions') }}</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200">
          <tr v-for="s in solutions" :key="s.id || s.slug">
            <td class="px-4 py-3 text-sm">{{ s.title }}</td>
            <td class="px-4 py-3 text-sm text-gray-500">{{ s.category || '—' }}</td>
            <td class="px-4 py-3 text-sm text-gray-500">{{ s.moq || '—' }}</td>
            <td class="px-4 py-3 text-sm space-x-2">
              <button @click="openSolutionForm(s)" class="text-orange-600 hover:underline text-xs">{{ t('admin.oem.edit') }}</button>
              <button @click="deleteSolution(s.id || s.slug)" class="text-red-600 hover:underline text-xs">{{ t('admin.oem.delete') }}</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Projects Tab -->
    <div v-if="activeTab === 'projects'">
      <div v-if="projectsLoading" class="text-center py-10 text-gray-500">{{ t('admin.oem.loading') }}</div>
      <div v-else-if="!projects.length" class="text-center py-10 text-gray-500">{{ t('admin.oem.no_projects') }}</div>
      <table v-else class="min-w-full bg-white rounded-lg shadow">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.oem.col_product') }}</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.oem.col_status') }}</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.oem.col_user') }}</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.oem.col_created') }}</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.oem.col_actions') }}</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200">
          <tr v-for="p in projects" :key="p.id">
            <td class="px-4 py-3 text-sm">{{ p.productName }}</td>
            <td class="px-4 py-3 text-sm">
              <span :class="statusClass(p.status)" class="px-2 py-0.5 rounded-full text-xs font-medium">{{ enumLabel('oem_status', p.status) }}</span>
            </td>
            <td class="px-4 py-3 text-sm text-gray-500">{{ p.userId }}</td>
            <td class="px-4 py-3 text-sm text-gray-500">{{ new Date(p.createdAt).toLocaleDateString() }}</td>
            <td class="px-4 py-3 text-sm space-x-2">
              <select @change="updateProjectStatus(p.id, ($event.target as HTMLSelectElement).value)" class="text-xs border rounded px-1 py-0.5">
                <option value="">{{ t('admin.oem.status_default') }}</option>
                <option value="inquiry">{{ enumLabel('oem_status', 'inquiry') }}</option>
                <option value="sampling">{{ enumLabel('oem_status', 'sampling') }}</option>
                <option value="formulation">{{ enumLabel('oem_status', 'formulation') }}</option>
                <option value="quotation">{{ enumLabel('oem_status', 'quotation') }}</option>
                <option value="contract">{{ enumLabel('oem_status', 'contract') }}</option>
                <option value="production">{{ enumLabel('oem_status', 'production') }}</option>
                <option value="delivery">{{ enumLabel('oem_status', 'delivery') }}</option>
                <option value="completed">{{ enumLabel('oem_status', 'completed') }}</option>
                <option value="cancelled">{{ enumLabel('oem_status', 'cancelled') }}</option>
              </select>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Flow Form Modal -->
    <Teleport to="body">
      <div v-if="showFlowForm" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" role="dialog" aria-modal="true" @click.self="showFlowForm = false">
        <div class="bg-white rounded-lg max-w-md w-full max-h-[90vh] overflow-y-auto shadow-xl">
          <div class="flex justify-between items-center px-6 py-4 border-b">
            <h3 class="text-lg font-semibold">{{ flowForm.id ? t('admin.oem.edit_flow') : t('admin.oem.new_flow_title') }}</h3>
            <button @click="showFlowForm = false" class="text-gray-400 hover:text-gray-600" :aria-label="t('close')">&times;</button>
          </div>
          <div class="px-6 py-4 space-y-3">
            <div>
              <label class="block text-sm font-medium mb-1">{{ t('admin.oem.label_title') }}</label>
              <input v-model="flowForm.title" class="w-full border rounded-md px-3 py-2 text-sm" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">{{ t('admin.oem.label_type') }}</label>
              <select v-model="flowForm.type" class="w-full border rounded-md px-3 py-2 text-sm">
                <option value="quick_odm">{{ t('admin.oem.type_quick_odm') }}</option>
                <option value="full_oem">{{ t('admin.oem.type_full_oem') }}</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">{{ t('admin.oem.label_description') }}</label>
              <textarea v-model="flowForm.description" rows="2" class="w-full border rounded-md px-3 py-2 text-sm"></textarea>
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">{{ t('admin.oem.label_timeline') }}</label>
              <input v-model="flowForm.timeline" class="w-full border rounded-md px-3 py-2 text-sm" />
            </div>
          </div>
          <div class="flex justify-end gap-3 px-6 py-4 border-t">
            <button @click="showFlowForm = false" class="px-4 py-2 text-sm border rounded-md">{{ t('admin.oem.cancel') }}</button>
            <button @click="saveFlow" :disabled="!flowForm.title.trim()" class="px-4 py-2 text-sm bg-orange-500 text-white rounded-md hover:bg-orange-600 disabled:opacity-50">{{ t('admin.oem.save') }}</button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Solution Form Modal -->
    <Teleport to="body">
      <div v-if="showSolutionForm" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" role="dialog" aria-modal="true" @click.self="showSolutionForm = false">
        <div class="bg-white rounded-lg max-w-md w-full max-h-[90vh] overflow-y-auto shadow-xl">
          <div class="flex justify-between items-center px-6 py-4 border-b">
            <h3 class="text-lg font-semibold">{{ solutionForm.id ? t('admin.oem.edit_solution') : t('admin.oem.new_solution_title') }}</h3>
            <button @click="showSolutionForm = false" class="text-gray-400 hover:text-gray-600" :aria-label="t('close')">&times;</button>
          </div>
          <div class="px-6 py-4 space-y-3">
            <div>
              <label class="block text-sm font-medium mb-1">{{ t('admin.oem.label_title') }}</label>
              <input v-model="solutionForm.title" class="w-full border rounded-md px-3 py-2 text-sm" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">{{ t('admin.oem.label_category') }}</label>
              <input v-model="solutionForm.category" class="w-full border rounded-md px-3 py-2 text-sm" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">{{ t('admin.oem.label_description') }}</label>
              <textarea v-model="solutionForm.description" rows="2" class="w-full border rounded-md px-3 py-2 text-sm"></textarea>
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">{{ t('admin.oem.label_thumbnail') }}</label>
              <input v-model="solutionForm.thumbnail" class="w-full border rounded-md px-3 py-2 text-sm" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">{{ t('admin.oem.label_moq') }}</label>
              <input v-model.number="solutionForm.moq" type="number" class="w-full border rounded-md px-3 py-2 text-sm" />
            </div>
          </div>
          <div class="flex justify-end gap-3 px-6 py-4 border-t">
            <button @click="showSolutionForm = false" class="px-4 py-2 text-sm border rounded-md">{{ t('admin.oem.cancel') }}</button>
            <button @click="saveSolution" :disabled="!solutionForm.title.trim()" class="px-4 py-2 text-sm bg-orange-500 text-white rounded-md hover:bg-orange-600 disabled:opacity-50">{{ t('admin.oem.save') }}</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: ['auth'] })

const api = useApi()
const { t } = useI18n()
const { enumLabel } = useDisplay()

const activeTab = ref('flows')
const tabs = computed(() => [
  { key: 'flows', label: t('admin.oem.tab_flows') },
  { key: 'solutions', label: t('admin.oem.tab_solutions') },
  { key: 'projects', label: t('admin.oem.tab_projects') },
])

const flows = ref<any[]>([])
const flowsLoading = ref(true)
const showFlowForm = ref(false)
const flowForm = reactive<any>({ id: '', title: '', type: 'quick_odm', description: '', timeline: '' })

const solutions = ref<any[]>([])
const solutionsLoading = ref(true)
const showSolutionForm = ref(false)
const solutionForm = reactive<any>({ id: '', title: '', category: '', description: '', thumbnail: '', moq: 500 })

const projects = ref<any[]>([])
const projectsLoading = ref(true)

const loadFlows = async () => {
  try { flows.value = await api.adminGetOemFlows() } catch {}
  flowsLoading.value = false
}
const loadSolutions = async () => {
  try { solutions.value = await api.adminGetOemSolutions() } catch {}
  solutionsLoading.value = false
}
const loadProjects = async () => {
  try {
    const res = await api.adminGetOemProjects()
    projects.value = res?.data || []
  } catch {}
  projectsLoading.value = false
}

onMounted(() => {
  loadFlows()
  loadSolutions()
  loadProjects()
})

const openFlowForm = (flow?: any) => {
  if (flow) {
    flowForm.id = flow.id
    flowForm.title = flow.title
    flowForm.type = flow.type || 'quick_odm'
    flowForm.description = flow.description || ''
    flowForm.timeline = flow.timeline || ''
  } else {
    flowForm.id = ''
    flowForm.title = ''
    flowForm.type = 'quick_odm'
    flowForm.description = ''
    flowForm.timeline = ''
  }
  showFlowForm.value = true
}

const saveFlow = async () => {
  try {
    if (flowForm.id) {
      await api.adminUpdateOemFlow(flowForm.id, { title: flowForm.title, type: flowForm.type, description: flowForm.description, timeline: flowForm.timeline })
    } else {
      await api.adminCreateOemFlow({ title: flowForm.title, type: flowForm.type, description: flowForm.description, timeline: flowForm.timeline })
    }
    showFlowForm.value = false
    await loadFlows()
  } catch {}
}

const deleteFlow = async (id: string) => {
  if (!confirm(t('admin.oem.confirm_delete_flow'))) return
  try {
    await api.adminDeleteOemFlow(id)
    await loadFlows()
  } catch {}
}

const openSolutionForm = (s?: any) => {
  if (s) {
    solutionForm.id = s.id || s.slug
    solutionForm.title = s.title
    solutionForm.category = s.category || ''
    solutionForm.description = s.description || ''
    solutionForm.thumbnail = s.thumbnail || ''
    solutionForm.moq = s.moq || 500
  } else {
    solutionForm.id = ''
    solutionForm.title = ''
    solutionForm.category = ''
    solutionForm.description = ''
    solutionForm.thumbnail = ''
    solutionForm.moq = 500
  }
  showSolutionForm.value = true
}

const saveSolution = async () => {
  try {
    if (solutionForm.id) {
      await api.adminUpdateOemSolution(solutionForm.id, {
        title: solutionForm.title, category: solutionForm.category,
        description: solutionForm.description, thumbnail: solutionForm.thumbnail, moq: solutionForm.moq
      })
    } else {
      await api.adminCreateOemSolution({
        title: solutionForm.title, category: solutionForm.category,
        description: solutionForm.description, thumbnail: solutionForm.thumbnail, moq: solutionForm.moq
      })
    }
    showSolutionForm.value = false
    await loadSolutions()
  } catch {}
}

const deleteSolution = async (id: string) => {
  if (!confirm(t('admin.oem.confirm_delete_solution'))) return
  try {
    await api.adminDeleteOemSolution(id)
    await loadSolutions()
  } catch {}
}

const updateProjectStatus = async (id: string, status: string) => {
  if (!status) return
  try {
    await api.adminUpdateOemStatus(id, { status })
    await loadProjects()
  } catch {}
}

const statusClass = (s: string) => {
  const map: Record<string, string> = {
    inquiry: 'bg-gray-100 text-gray-700',
    sampling: 'bg-blue-100 text-blue-700',
    formulation: 'bg-yellow-100 text-yellow-700',
    quotation: 'bg-orange-100 text-orange-700',
    contract: 'bg-purple-100 text-purple-700',
    production: 'bg-cyan-100 text-cyan-700',
    delivery: 'bg-green-100 text-green-700',
    completed: 'bg-emerald-100 text-emerald-700',
    cancelled: 'bg-red-100 text-red-700',
  }
  return map[s] || 'bg-gray-100 text-gray-600'
}
</script>
