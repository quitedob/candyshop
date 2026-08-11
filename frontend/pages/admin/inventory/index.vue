<template>
  <div>
    <div class="sm:flex sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">{{ t('admin.inventory.title') }}</h1>
        <p class="mt-1 text-sm text-gray-600">{{ t('admin.inventory.description') }}</p>
      </div>
      <div class="mt-4 sm:mt-0 flex items-center gap-3">
        <button v-if="inventoryTab === 'stock'" @click="exportSelected" :disabled="selectedIds.size === 0" class="inline-flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed transition-colors">
          <Icon name="heroicons:arrow-down-tray" class="h-4 w-4" aria-hidden="true" />
          {{ selectedIds.size > 0 ? t('admin.inventory.export_selected') + ` (${selectedIds.size})` : t('admin.inventory.export_selected') }}
        </button>
        <button v-if="inventoryTab === 'stock'" @click="openImportModal" class="inline-flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 transition-colors">
          <Icon name="heroicons:arrow-up-tray" class="h-4 w-4" aria-hidden="true" />
          {{ t('admin.inventory.import_xlsx') }}
        </button>
        <button v-if="inventoryTab === 'stock'" @click="downloadTemplate" class="inline-flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 transition-colors">
          <Icon name="heroicons:document-arrow-down" class="h-4 w-4" aria-hidden="true" />
          {{ t('admin.inventory.import_download_template') }}
        </button>
        <button v-if="inventoryTab === 'stock'" @click="openAdjustmentModal()" class="inline-flex items-center gap-2 px-4 py-2 bg-orange-600 text-white rounded-lg text-sm font-medium hover:bg-orange-700 transition-colors">
          <Icon name="heroicons:adjustments-horizontal" class="h-4 w-4" aria-hidden="true" />
          {{ t('admin.inventory.adjust_stock') }}
        </button>
      </div>
    </div>

    <!-- Tabs: stock / warehouses / transfers -->
    <div class="mt-6 border-b border-gray-200">
      <nav class="-mb-px flex gap-6">
        <button type="button" :class="inventoryTab === 'stock' ? 'border-orange-500 text-orange-600' : 'border-transparent text-gray-500'" class="whitespace-nowrap border-b-2 py-3 text-sm font-medium" @click="inventoryTab = 'stock'">{{ t('admin.inventory.tab_stock') }}</button>
        <button v-if="enableMultiWarehouse" type="button" :class="inventoryTab === 'warehouses' ? 'border-orange-500 text-orange-600' : 'border-transparent text-gray-500'" class="whitespace-nowrap border-b-2 py-3 text-sm font-medium" @click="inventoryTab = 'warehouses'; fetchWarehouses()">{{ t('admin.inventory.tab_warehouses') }}</button>
        <button v-if="enableMultiWarehouse" type="button" :class="inventoryTab === 'transfers' ? 'border-orange-500 text-orange-600' : 'border-transparent text-gray-500'" class="whitespace-nowrap border-b-2 py-3 text-sm font-medium" @click="inventoryTab = 'transfers'; fetchTransfers()">{{ t('admin.inventory.tab_transfers') }}</button>
      </nav>
    </div>

    <!-- Warehouses tab（多仓库功能暂隐藏） -->
    <div v-if="enableMultiWarehouse && inventoryTab === 'warehouses'" class="mt-6 space-y-4">
      <div class="flex justify-end">
        <button type="button" class="rounded-lg bg-orange-600 px-4 py-2 text-sm text-white hover:bg-orange-700" @click="openWarehouseModal()">{{ t('admin.inventory.add_warehouse') }}</button>
      </div>
      <div class="overflow-hidden rounded-xl border border-gray-200 bg-white">
        <table class="min-w-full divide-y divide-gray-200 text-sm">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-4 py-3 text-start font-semibold text-gray-600">{{ t('admin.inventory.wh_code') }}</th>
              <th class="px-4 py-3 text-start font-semibold text-gray-600">{{ t('admin.inventory.wh_name') }}</th>
              <th class="px-4 py-3 text-start font-semibold text-gray-600">{{ t('admin.inventory.wh_country') }}</th>
              <th class="px-4 py-3 text-start font-semibold text-gray-600">{{ t('admin.inventory.col_actions') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200">
            <tr v-if="warehousesPending"><td colspan="4" class="px-4 py-8 text-center text-gray-500">{{ t('admin.inventory.loading') }}</td></tr>
            <tr v-else-if="!warehouses.length"><td colspan="4" class="px-4 py-8 text-center text-gray-500">{{ t('admin.inventory.no_warehouses') }}</td></tr>
            <tr v-else v-for="wh in warehouses" :key="wh.id">
              <td class="px-4 py-3 font-mono">{{ wh.code }}</td>
              <td class="px-4 py-3">{{ wh.name }}</td>
              <td class="px-4 py-3">{{ wh.country || '-' }}</td>
              <td class="px-4 py-3">
                <button type="button" class="text-orange-600 hover:text-orange-900 me-3" @click="openWarehouseModal(wh)">{{ t('admin.inventory.edit') }}</button>
                <button type="button" class="text-gray-600 hover:text-gray-900" @click="openWarehouseStockModal(wh)">{{ t('admin.inventory.update_stock') }}</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Transfers tab -->
    <div v-else-if="enableMultiWarehouse && inventoryTab === 'transfers'" class="mt-6 space-y-4">
      <div class="flex justify-end">
        <button type="button" class="rounded-lg bg-orange-600 px-4 py-2 text-sm text-white hover:bg-orange-700" @click="openTransferModal">{{ t('admin.inventory.create_transfer') }}</button>
      </div>
      <div class="overflow-hidden rounded-xl border border-gray-200 bg-white">
        <table class="min-w-full divide-y divide-gray-200 text-sm">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-4 py-3 text-start font-semibold text-gray-600">{{ t('admin.inventory.transfer_no') }}</th>
              <th class="px-4 py-3 text-start font-semibold text-gray-600">{{ t('admin.inventory.from_wh') }}</th>
              <th class="px-4 py-3 text-start font-semibold text-gray-600">{{ t('admin.inventory.to_wh') }}</th>
              <th class="px-4 py-3 text-start font-semibold text-gray-600">{{ t('admin.inventory.col_status') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200">
            <tr v-if="transfersPending"><td colspan="4" class="px-4 py-8 text-center text-gray-500">{{ t('admin.inventory.loading') }}</td></tr>
            <tr v-else-if="!transfers.length"><td colspan="4" class="px-4 py-8 text-center text-gray-500">{{ t('admin.inventory.no_transfers') }}</td></tr>
            <tr v-else v-for="tr in transfers" :key="tr.id">
              <td class="px-4 py-3 font-mono">{{ tr.transferNumber || tr.id }}</td>
              <td class="px-4 py-3">{{ tr.fromWarehouseId }}</td>
              <td class="px-4 py-3">{{ tr.toWarehouseId }}</td>
              <td class="px-4 py-3">{{ tr.status }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <template v-else>
    <!-- Stats Cards -->
    <div class="mt-6 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-5">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500">{{ t('admin.inventory.total_products') }}</p>
            <p class="mt-1 text-2xl font-bold text-gray-900">{{ displayPagination?.total || items.length }}</p>
          </div>
          <div class="h-12 w-12 rounded-lg bg-orange-50 flex items-center justify-center">
            <Icon name="heroicons:cube" class="h-6 w-6 text-orange-600" aria-hidden="true" />
          </div>
        </div>
      </div>
      <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-5">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500">{{ t('admin.inventory.low_stock') }}</p>
            <p class="mt-1 text-2xl font-bold text-red-600">{{ stats.lowStock }}</p>
          </div>
          <div class="h-12 w-12 rounded-lg bg-red-50 flex items-center justify-center">
            <Icon name="heroicons:exclamation-triangle" class="h-6 w-6 text-red-600" aria-hidden="true" />
          </div>
        </div>
      </div>
      <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-5">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500">{{ t('admin.inventory.out_of_stock') }}</p>
            <p class="mt-1 text-2xl font-bold text-orange-600">{{ stats.outOfStock }}</p>
          </div>
          <div class="h-12 w-12 rounded-lg bg-orange-50 flex items-center justify-center">
            <Icon name="heroicons:x-circle" class="h-6 w-6 text-orange-600" aria-hidden="true" />
          </div>
        </div>
      </div>
      <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-5">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500">{{ t('admin.inventory.total_value') }}</p>
            <p class="mt-1 text-2xl font-bold text-emerald-600">{{ cur(stats.currency) }} {{ formatNumber(stats.totalValue) }}</p>
          </div>
          <div class="h-12 w-12 rounded-lg bg-emerald-50 flex items-center justify-center">
            <Icon name="heroicons:currency-dollar" class="h-6 w-6 text-emerald-600" aria-hidden="true" />
          </div>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="mt-6 bg-white rounded-xl shadow-sm border border-gray-200 p-4">
      <div class="flex flex-wrap items-center gap-4">
        <div class="flex-1 min-w-[200px]">
          <div class="relative">
            <label for="inventory-search" class="sr-only">{{ t('admin.inventory.search') }}</label>
            <Icon name="heroicons:magnifying-glass" class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" aria-hidden="true" />
            <input id="inventory-search" v-model="searchQuery" name="search" type="text" autocomplete="off" :placeholder="t('admin.inventory.search')" class="w-full ps-10 pe-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-orange-500" />
          </div>
        </div>
        <label for="inventory-stockFilter" class="sr-only">{{ t('admin.inventory.filter_stock') }}</label>
        <select id="inventory-stockFilter" v-model="stockFilter" name="stockFilter" autocomplete="off" class="px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-orange-500">
          <option value="all">{{ t('admin.inventory.filter_all') }}</option>
          <option value="low">{{ t('admin.inventory.filter_low') }}</option>
          <option value="out">{{ t('admin.inventory.filter_out') }}</option>
          <option value="in">{{ t('admin.inventory.filter_in') }}</option>
        </select>
        <label for="inventory-categoryFilter" class="sr-only">{{ t('admin.inventory.filter_category') }}</label>
        <select id="inventory-categoryFilter" v-model="categoryFilter" name="categoryFilter" autocomplete="off" class="px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-orange-500">
          <option value="">{{ t('admin.inventory.all_categories') }}</option>
          <option v-for="cat in categoryOptions" :key="cat.slug" :value="cat.slug">{{ cat.label }}</option>
        </select>
        <button v-if="searchQuery || stockFilter !== 'all' || categoryFilter" @click="clearFilters" class="text-sm text-orange-600 hover:text-orange-800">
          {{ t('admin.inventory.clear_filters') }}
        </button>
      </div>
    </div>

    <!-- Bulk Actions Bar -->
    <div v-if="selectedIds.size > 0" class="mt-4 bg-orange-50 rounded-xl border border-orange-200 p-3 flex items-center justify-between">
      <span class="text-sm text-orange-800">{{ t('admin.inventory.selected_count', { count: selectedIds.size }) }}</span>
      <div class="flex items-center gap-2">
        <button @click="exportSelected" class="px-3 py-1.5 bg-orange-600 text-white rounded-lg text-sm font-medium hover:bg-orange-700 transition-colors">
          {{ t('admin.inventory.export_selected') }}
        </button>
        <button @click="openBatchEditModal" class="px-3 py-1.5 bg-orange-600 text-white rounded-lg text-sm font-medium hover:bg-orange-700 transition-colors">
          {{ t('admin.inventory.batch_edit') }}
        </button>
        <button @click="openBatchDeleteModal" class="px-3 py-1.5 bg-red-600 text-white rounded-lg text-sm font-medium hover:bg-red-700 transition-colors">
          {{ t('admin.inventory.batch_delete') }}
        </button>
        <button @click="clearSelection" class="px-3 py-1.5 border border-orange-300 text-orange-700 rounded-lg text-sm hover:bg-orange-100 transition-colors">
          {{ t('admin.inventory.clear_selection') }}
        </button>
      </div>
    </div>

    <!-- Inventory Table -->
    <div class="mt-6 bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-4 py-3 w-10">
                <input type="checkbox" name="selectAll" :checked="isAllSelected" @change="toggleSelectAll" class="h-4 w-4 rounded border-gray-300 text-orange-600 focus:ring-orange-500" />
              </th>
              <th class="px-6 py-3 text-start text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.inventory.col_product') }}</th>
              <th class="px-6 py-3 text-start text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.inventory.col_category') }}</th>
              <th class="px-6 py-3 text-start text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.inventory.col_sku') }}</th>
              <th class="px-6 py-3 text-end text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.inventory.col_stock') }}</th>
              <th class="px-6 py-3 text-end text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.inventory.col_moq') }}</th>
              <th class="px-6 py-3 text-end text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.inventory.col_unit_value') }}</th>
              <th class="px-6 py-3 text-end text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.inventory.col_total_value') }}</th>
              <th class="px-6 py-3 text-start text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.inventory.col_status') }}</th>
              <th class="px-6 py-3 text-start text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.inventory.col_actions') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200 bg-white">
            <tr v-if="pending">
              <td colspan="10" class="px-6 py-10 text-center text-sm text-gray-500">{{ t('admin.inventory.loading') }}</td>
            </tr>
            <tr v-else-if="error">
              <td colspan="10" class="px-6 py-10 text-center text-sm text-red-600">{{ error }}</td>
            </tr>
            <tr v-else-if="filteredItems.length === 0">
              <td colspan="10" class="px-6 py-10 text-center text-sm text-gray-500">{{ t('admin.inventory.no_data') }}</td>
            </tr>
            <tr v-else v-for="item in filteredItems" :key="item.id" :class="['hover:bg-gray-50 transition-colors', selectedIds.has(item.id) ? 'bg-orange-50' : '']">
              <td class="px-4 py-4">
                <input type="checkbox" name="selectItem" :checked="selectedIds.has(item.id)" @change="toggleSelect(item.id)" class="h-4 w-4 rounded border-gray-300 text-orange-600 focus:ring-orange-500" />
              </td>
              <td class="px-6 py-4">
                <div class="flex items-center gap-3">
                  <div class="h-10 w-10 rounded-lg bg-gradient-to-br from-orange-50 to-amber-50 flex items-center justify-center overflow-hidden">
                    <img v-if="item.thumbnail" :src="item.thumbnail" class="h-full w-full object-cover" width="40" height="40" alt="" />
                    <Icon v-else name="heroicons:cube" class="h-5 w-5 text-orange-300" aria-hidden="true" />
                  </div>
                  <div>
                    <div class="font-medium text-gray-900">{{ tField(item, 'name') }}</div>
                    <div class="text-xs text-gray-500">{{ item.slug }}</div>
                  </div>
                </div>
              </td>
              <td class="px-6 py-4 text-sm text-gray-600">{{ categoryLabel(item) }}</td>
              <td class="px-6 py-4 text-sm text-gray-600 font-mono">{{ item.id?.substring(0, 8).toUpperCase() || '-' }}</td>
              <td class="px-6 py-4 text-sm text-end">
                <div class="flex items-center justify-end gap-2">
                  <span :class="stockStatusClass(item.stockQuantity, item.moq)" class="font-semibold">{{ item.stockQuantity || 0 }}</span>
                  <button @click="openAdjustmentModal(item)" class="p-1 text-orange-600 hover:bg-orange-50 rounded transition-colors" :aria-label="t('admin.inventory.adjust_stock')">
                    <Icon name="heroicons:pencil" class="h-4 w-4" aria-hidden="true" />
                  </button>
                </div>
              </td>
              <td class="px-6 py-4 text-sm text-end text-gray-600">{{ item.moq || 0 }}</td>
              <td class="px-6 py-4 text-sm text-end text-gray-600">{{ cur(item.basePrice) }} {{ formatNumber(item.basePrice || 0) }}</td>
              <td class="px-6 py-4 text-sm text-end font-medium text-gray-900">{{ cur(item.basePrice) }} {{ formatNumber((item.stockQuantity || 0) * (item.basePrice || 0)) }}</td>
              <td class="px-6 py-4">
                <span :class="stockBadgeClass(item.stockQuantity, item.moq)" class="inline-flex rounded-full px-2.5 py-0.5 text-xs font-semibold">
                  {{ stockStatusLabel(item.stockQuantity, item.moq) }}
                </span>
              </td>
              <td class="px-6 py-4 text-sm">
                <button @click="viewHistory(item)" class="text-orange-600 hover:text-orange-900 me-3">{{ t('admin.inventory.history') }}</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div v-if="displayPagination" class="px-6 py-4 border-t border-gray-200 flex items-center justify-between">
        <div class="text-sm text-gray-600">
          {{ t('admin.inventory.showing', { from: ((page - 1) * pageSize) + 1, to: Math.min(page * pageSize, displayPagination.total), total: displayPagination.total }) }}
        </div>
        <div class="flex items-center gap-2">
          <button @click="prevPage" :disabled="page <= 1" class="px-3 py-1.5 border border-gray-200 rounded-lg text-sm hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors">
            {{ t('admin.inventory.previous') }}
          </button>
          <button @click="nextPage" :disabled="page >= displayPagination.totalPages" class="px-3 py-1.5 border border-gray-200 rounded-lg text-sm hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors">
            {{ t('admin.inventory.next') }}
          </button>
        </div>
      </div>
    </div>
    </template>

    <!-- Import XLSX Modal -->
    <div v-if="showImportModal" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-screen items-center justify-center px-4 pb-20 pt-4 text-center sm:block sm:p-0">
        <button type="button" class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity w-full border-0 cursor-pointer" @click="closeImportModal" :aria-label="t('close')"></button>
        <div class="inline-block w-full transform overflow-hidden rounded-xl bg-white text-start align-bottom shadow-xl sm:my-8 sm:max-w-5xl sm:align-middle">
          <div class="bg-gradient-to-r from-orange-500 to-amber-600 px-6 py-4 flex items-center justify-between">
            <div>
              <h3 class="text-lg font-semibold text-white">{{ t('admin.inventory.import_title') }}</h3>
              <p class="text-sm text-orange-100 mt-0.5">{{ t('admin.inventory.import_drop_zone') }}</p>
            </div>
            <button @click="closeImportModal" class="text-white hover:text-gray-200" :aria-label="t('close')">
              <Icon name="heroicons:x-mark" class="h-5 w-5" aria-hidden="true" />
            </button>
          </div>

          <div class="p-6">
            <!-- File Upload -->
            <div v-if="!importPreview" class="flex flex-col items-center gap-4">
              <div class="w-full max-w-lg border-2 border-dashed border-gray-300 rounded-xl p-10 text-center hover:border-orange-400 transition-colors cursor-pointer" role="button" tabindex="0" :aria-label="t('admin.inventory.import_drop_zone')" @click="triggerFileInput" @keydown.enter.prevent="triggerFileInput" @keydown.space.prevent="triggerFileInput" @dragover.prevent @drop.prevent="handleFileDrop">
                <Icon name="heroicons:document-arrow-up" class="h-12 w-12 text-gray-400 mx-auto mb-3" aria-hidden="true" />
                <p class="text-sm text-gray-600">{{ t('admin.inventory.import_drop_zone') }}</p>
                <p class="text-xs text-gray-400 mt-1">.xlsx</p>
                <input ref="fileInput" id="inventory-xlsx-file" name="xlsxFile" type="file" accept=".xlsx" autocomplete="off" class="hidden" @change="handleFileSelect" />
              </div>
              <div v-if="importError" class="text-sm text-red-600">{{ importError }}</div>
              <div v-if="importLoading" class="flex items-center gap-2 text-sm text-gray-600">
                <Icon name="heroicons:arrow-path" class="h-4 w-4 animate-spin" aria-hidden="true" />
                {{ t('admin.inventory.import_parsing') }}
              </div>
            </div>

            <!-- Preview -->
            <div v-else>
              <div class="flex items-center justify-between mb-4">
                <div>
                  <h4 class="font-semibold text-gray-900">{{ t('admin.inventory.import_preview') }}</h4>
                  <p class="text-sm text-gray-500">{{ importPreview.totalRows }} rows | {{ importPreview.headers.length }} columns</p>
                </div>
                <div class="flex items-center gap-3">
                  <button @click="resetImport" class="px-3 py-1.5 border border-gray-300 rounded-lg text-sm hover:bg-gray-50 transition-colors">
                    {{ t('admin.inventory.cancel') }}
                  </button>
                  <button @click="applyImport" :disabled="importApplying" class="px-4 py-1.5 bg-orange-600 text-white rounded-lg text-sm font-medium hover:bg-orange-700 disabled:opacity-50 transition-colors flex items-center gap-2">
                    <Icon v-if="importApplying" name="heroicons:arrow-path" class="h-4 w-4 animate-spin" aria-hidden="true" />
                    {{ importApplying ? t('admin.inventory.import_applying') : t('admin.inventory.import_apply') }}
                  </button>
                </div>
              </div>

              <!-- AI mapping banner -->
              <div v-if="importPreview.aiMapping && Object.keys(importPreview.aiMapping).length > 0" class="mb-3 bg-blue-50 border border-blue-200 rounded-lg p-3">
                <p class="text-sm font-medium text-blue-800 inline-flex items-center">
                  {{ t('admin.inventory.import_ai_mapping') }}
                  <AiHelpHint topic="inventory_import_ai" size="sm" />
                </p>
                <p class="text-xs text-blue-600 mt-1">{{ aiMappingSummary }}</p>
              </div>

              <!-- Image detection banner -->
              <div v-if="importPreview.imageColumns && importPreview.imageColumns.length > 0" class="mb-3 bg-emerald-50 border border-emerald-200 rounded-lg p-3">
                <div class="flex items-center gap-2">
                  <Icon name="heroicons:photo" class="h-5 w-5 text-emerald-600" aria-hidden="true" />
                  <p class="text-sm text-emerald-800">{{ t('admin.inventory.import_image_detected') }} ({{ importPreview.imageColumns.length }} column(s))</p>
                </div>
              </div>

              <!-- Column mapping -->
              <div class="mb-4 bg-gray-50 rounded-lg p-3">
                <p class="text-sm font-medium text-gray-700 mb-2">{{ t('admin.inventory.import_column_mapping') }}</p>
                <div class="flex flex-wrap gap-2">
                  <span v-for="(colIdx, field) in activeMapping" :key="field" class="inline-flex items-center gap-1 px-2.5 py-1 bg-white border border-gray-200 rounded-md text-xs">
                    <span class="text-gray-500">{{ field }}</span>
                    <span class="text-gray-300">→</span>
                    <span class="text-orange-700 font-medium">{{ importPreview.headers[parseInt(colIdx)] }}</span>
                  </span>
                </div>
              </div>

              <!-- Data Preview Table -->
              <div class="max-h-80 overflow-auto border border-gray-200 rounded-lg">
                <table class="min-w-full divide-y divide-gray-200 text-xs">
                  <thead class="bg-gray-50 sticky top-0">
                    <tr>
                      <th class="px-3 py-2 text-start font-semibold text-gray-600">#</th>
                      <th v-for="(h, i) in importPreview.headers" :key="i" class="px-3 py-2 text-start font-semibold text-gray-600 whitespace-nowrap">
                        <div class="flex items-center gap-1">
                          {{ h }}
                          <Icon v-if="importPreview.imageColumns?.includes(i)" name="heroicons:photo" class="h-3 w-3 text-emerald-600" :title="$t('inventory.image_column')" aria-hidden="true" />
                        </div>
                      </th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-gray-100">
                    <tr v-for="(row, ri) in importPreview.rows.slice(0, 20)" :key="ri" class="hover:bg-gray-50">
                      <td class="px-3 py-1.5 text-gray-400">{{ Number(ri) + 1 }}</td>
                      <td v-for="(cell, ci) in row" :key="ci" class="px-3 py-1.5 text-gray-700 max-w-[200px] truncate">
                        <template v-if="importPreview.imageColumns?.includes(ci) && isImageURL(cell)">
                          <div class="flex items-center gap-1.5">
                            <img :src="cell" class="h-6 w-6 rounded object-cover" @error="hideBrokenImage" />
                            <span class="text-emerald-600 truncate">{{ cell }}</span>
                          </div>
                        </template>
                        <template v-else>
                          {{ cell }}
                        </template>
                      </td>
                    </tr>
                  </tbody>
                </table>
                <div v-if="importPreview.rows.length > 20" class="px-3 py-2 text-xs text-gray-400 bg-gray-50 border-t border-gray-200 text-center">
                  {{ $t('admin.inventory.import_preview_more', { n: importPreview.rows.length - 20 }) }}
                </div>
              </div>
            </div>

            <!-- Import Result -->
            <div v-if="importResult" class="mt-4 bg-gray-50 rounded-lg p-4">
              <p class="text-sm font-medium text-gray-900">{{ t('admin.inventory.import_result', { created: importResult.created, updated: importResult.updated, errors: importResult.errors }) }}</p>
              <ul v-if="importResult.errorMessages?.length" class="mt-2 text-xs text-red-600 space-y-1">
                <li v-for="(msg, i) in importResult.errorMessages" :key="i">{{ msg }}</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Batch Edit Modal -->
    <div v-if="showBatchEditModal" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-screen items-center justify-center px-4 pb-20 pt-4 text-center sm:block sm:p-0">
        <button type="button" class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity w-full border-0 cursor-pointer" @click="closeBatchEditModal" :aria-label="t('close')"></button>
        <div class="inline-block w-full transform overflow-hidden rounded-xl bg-white text-start align-bottom shadow-xl sm:my-8 sm:max-w-lg sm:align-middle">
          <div class="bg-gradient-to-r from-orange-500 to-amber-600 px-6 py-4 flex items-center justify-between">
            <div>
              <h3 class="text-lg font-semibold text-white">{{ t('admin.inventory.batch_edit_title') }}</h3>
              <p class="text-sm text-orange-100 mt-0.5">{{ t('admin.inventory.batch_edit_selected_count', { count: selectedIds.size }) }}</p>
            </div>
            <button @click="closeBatchEditModal" class="text-white hover:text-gray-200" :aria-label="t('close')">
              <Icon name="heroicons:x-mark" class="h-5 w-5" aria-hidden="true" />
            </button>
          </div>
          <div class="p-6 space-y-4">
            <div>
              <label for="inventory-batchField" class="block text-sm font-medium text-gray-700">{{ t('admin.inventory.batch_edit_field') }}</label>
              <select id="inventory-batchField" v-model="batchEditField" name="batchEditField" autocomplete="off" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option value="">{{ t('admin.inventory.batch_edit_select_field') }}</option>
                <option value="category">{{ t('admin.inventory.batch_edit_field_category') }}</option>
                <option value="categorySlug">{{ t('admin.inventory.batch_edit_field_categorySlug') }}</option>
                <option value="status">{{ t('admin.inventory.batch_edit_field_status') }}</option>
                <option value="stockQuantity">{{ t('admin.inventory.batch_edit_field_stockQuantity') }}</option>
                <option value="basePrice">{{ t('admin.inventory.batch_edit_field_basePrice') }}</option>
                <option value="moq">{{ t('admin.inventory.batch_edit_field_moq') }}</option>
                <option value="leadTime">{{ t('admin.inventory.batch_edit_field_leadTime') }}</option>
                <option value="halalCertified">{{ t('admin.inventory.batch_edit_field_halalCertified') }}</option>
                <option value="oemAvailable">{{ t('admin.inventory.batch_edit_field_oemAvailable') }}</option>
                <option value="hsCode">{{ t('admin.inventory.batch_edit_field_hsCode') }}</option>
                <option value="shelfLife">{{ t('admin.inventory.batch_edit_field_shelfLife') }}</option>
                <option value="storage">{{ t('admin.inventory.batch_edit_field_storage') }}</option>
              </select>
            </div>

            <div v-if="batchEditField">
              <label for="inventory-batchValue" class="block text-sm font-medium text-gray-700">{{ t('admin.inventory.batch_edit_value') }}</label>

              <!-- Status select -->
              <select v-if="batchEditField === 'status'" id="inventory-batchValue" v-model="batchEditValue" name="batchEditValue" autocomplete="off" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option value="active">{{ t('admin.users.status_active') }}</option>
                <option value="inactive">{{ t('admin.users.status_inactive') }}</option>
                <option value="draft">{{ t('admin.content.status_draft') }}</option>
              </select>

              <!-- Boolean toggles -->
              <div v-else-if="batchEditField === 'halalCertified' || batchEditField === 'oemAvailable'" class="mt-2 flex gap-4">
                <label class="flex items-center gap-2 cursor-pointer">
                  <input v-model="batchEditValue" name="batchBool" type="radio" :value="true" class="text-orange-600 focus:ring-orange-500" />
                  <span class="text-sm text-gray-700">{{ t('form.yes') }}</span>
                </label>
                <label class="flex items-center gap-2 cursor-pointer">
                  <input v-model="batchEditValue" name="batchBool" type="radio" :value="false" class="text-orange-600 focus:ring-orange-500" />
                  <span class="text-sm text-gray-700">{{ t('form.no') }}</span>
                </label>
              </div>

              <!-- Numeric inputs -->
              <input v-else-if="batchEditField === 'stockQuantity' || batchEditField === 'moq'" id="inventory-batchValue" v-model.number="batchEditValue" name="batchEditValue" type="number" min="0" autocomplete="off" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" />

              <!-- Price input -->
              <input v-else-if="batchEditField === 'basePrice'" id="inventory-batchValue" v-model.number="batchEditValue" name="batchEditValue" type="number" min="0" step="0.01" autocomplete="off" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" />

              <!-- Text inputs (category, categorySlug, leadTime, hsCode, shelfLife, storage) -->
              <input v-else id="inventory-batchValue" v-model="batchEditValue" name="batchEditValue" type="text" autocomplete="off" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>

            <div v-if="batchEditError" class="text-sm text-red-600">{{ batchEditError }}</div>
            <div v-if="batchEditResult" class="text-sm rounded-lg p-3" :class="batchEditResult.errors > 0 ? 'bg-orange-50 text-orange-800 border border-orange-200' : 'bg-emerald-50 text-emerald-800 border border-emerald-200'">
              {{ t('admin.inventory.batch_edit_result', { updated: batchEditResult.updated, errors: batchEditResult.errors }) }}
              <ul v-if="batchEditResult.errorMessages?.length" class="mt-1 text-xs space-y-0.5">
                <li v-for="(msg, i) in batchEditResult.errorMessages" :key="i">{{ msg }}</li>
              </ul>
            </div>
          </div>
          <div class="px-6 py-4 bg-gray-50 flex justify-end gap-3">
            <button @click="closeBatchEditModal" class="px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100 transition-colors">
              {{ t('admin.inventory.cancel') }}
            </button>
            <button @click="submitBatchEdit" :disabled="!batchEditField || batchEditValue === '' || batchEditValue === null || batchEditApplying" class="px-4 py-2 bg-orange-600 text-white rounded-lg text-sm font-medium hover:bg-orange-700 disabled:opacity-50 transition-colors flex items-center gap-2">
              <Icon v-if="batchEditApplying" name="heroicons:arrow-path" class="h-4 w-4 animate-spin" aria-hidden="true" />
              {{ batchEditApplying ? t('admin.inventory.batch_edit_applying') : t('admin.inventory.batch_edit_apply') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Batch Delete Confirmation Modal -->
    <div v-if="showBatchDeleteModal" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-screen items-center justify-center px-4 pb-20 pt-4 text-center sm:block sm:p-0">
        <button type="button" class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity w-full border-0 cursor-pointer" @click="closeBatchDeleteModal" :aria-label="t('close')"></button>
        <div class="inline-block w-full transform overflow-hidden rounded-xl bg-white text-start align-bottom shadow-xl sm:my-8 sm:max-w-md sm:align-middle">
          <div class="bg-gradient-to-r from-red-500 to-rose-600 px-6 py-4 flex items-center justify-between">
            <div>
              <h3 class="text-lg font-semibold text-white">{{ t('admin.inventory.batch_delete_title') }}</h3>
            </div>
            <button @click="closeBatchDeleteModal" class="text-white hover:text-gray-200" :aria-label="t('close')">
              <Icon name="heroicons:x-mark" class="h-5 w-5" aria-hidden="true" />
            </button>
          </div>
          <div class="p-6">
            <div class="flex items-center gap-3">
              <Icon name="heroicons:exclamation-triangle" class="h-8 w-8 text-red-500 flex-shrink-0" aria-hidden="true" />
              <p class="text-sm text-gray-700">{{ t('admin.inventory.batch_delete_confirm', { count: selectedIds.size }) }}</p>
            </div>
            <div v-if="batchDeleteError" class="mt-3 text-sm text-red-600">{{ batchDeleteError }}</div>
            <div v-if="batchDeleteResult" class="mt-3 text-sm rounded-lg p-3" :class="batchDeleteResult.errors > 0 ? 'bg-orange-50 text-orange-800 border border-orange-200' : 'bg-emerald-50 text-emerald-800 border border-emerald-200'">
              {{ t('admin.inventory.batch_delete_result', { deleted: batchDeleteResult.deleted, errors: batchDeleteResult.errors }) }}
              <ul v-if="batchDeleteResult.errorMessages?.length" class="mt-1 text-xs space-y-0.5">
                <li v-for="(msg, i) in batchDeleteResult.errorMessages" :key="i">{{ msg }}</li>
              </ul>
            </div>
          </div>
          <div class="px-6 py-4 bg-gray-50 flex justify-end gap-3">
            <button @click="closeBatchDeleteModal" class="px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100 transition-colors">
              {{ t('admin.inventory.cancel') }}
            </button>
            <button @click="submitBatchDelete" :disabled="batchDeleteApplying" class="px-4 py-2 bg-red-600 text-white rounded-lg text-sm font-medium hover:bg-red-700 disabled:opacity-50 transition-colors flex items-center gap-2">
              <Icon v-if="batchDeleteApplying" name="heroicons:arrow-path" class="h-4 w-4 animate-spin" aria-hidden="true" />
              {{ batchDeleteApplying ? t('admin.inventory.batch_delete_applying') : t('admin.inventory.batch_delete') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Stock Adjustment Modal -->
    <div v-if="showAdjustmentModal" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-screen items-end justify-center px-4 pb-20 pt-4 text-center sm:block sm:p-0">
        <button type="button" class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity w-full border-0 cursor-pointer" @click="closeAdjustmentModal" :aria-label="t('close')"></button>
        <div class="inline-block w-full transform overflow-hidden rounded-xl bg-white text-start align-bottom shadow-xl sm:my-8 sm:max-w-lg sm:align-middle">
          <div class="bg-gradient-to-r from-orange-500 to-amber-600 px-6 py-4">
            <h3 class="text-lg font-semibold text-white">{{ t('admin.inventory.adjust_title') }}</h3>
            <p class="text-sm text-orange-100 mt-0.5">{{ tField(adjustmentProduct, 'name') }}</p>
          </div>
          <div class="p-6 space-y-4">
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label for="inventory-currentStock" class="block text-sm font-medium text-gray-700">{{ t('admin.inventory.current_stock') }}</label>
                <p class="mt-1 text-2xl font-bold text-gray-900">{{ adjustmentProduct?.stockQuantity || 0 }}</p>
              </div>
              <div>
                <label for="inventory-newStock" class="block text-sm font-medium text-gray-700">{{ t('admin.inventory.new_stock') }}</label>
                <input id="inventory-newStock" v-model.number="newStock" name="newStock" type="number" min="0" autocomplete="off" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500" />
              </div>
            </div>
            <div>
              <label for="inventory-adjustType" class="block text-sm font-medium text-gray-700">{{ t('admin.inventory.adjustment_type') }}</label>
              <div class="mt-2 flex gap-4">
                <label class="flex items-center gap-2 cursor-pointer">
                  <input id="inventory-adjustType-set" v-model="adjustmentType" name="adjustmentType" type="radio" value="set" class="text-orange-600 focus:ring-orange-500" />
                  <span class="text-sm text-gray-700">{{ t('admin.inventory.type_set') }}</span>
                </label>
                <label class="flex items-center gap-2 cursor-pointer">
                  <input id="inventory-adjustType-add" v-model="adjustmentType" name="adjustmentType" type="radio" value="add" class="text-orange-600 focus:ring-orange-500" />
                  <span class="text-sm text-gray-700">{{ t('admin.inventory.type_add') }}</span>
                </label>
                <label class="flex items-center gap-2 cursor-pointer">
                  <input id="inventory-adjustType-subtract" v-model="adjustmentType" name="adjustmentType" type="radio" value="subtract" class="text-orange-600 focus:ring-orange-500" />
                  <span class="text-sm text-gray-700">{{ t('admin.inventory.type_subtract') }}</span>
                </label>
              </div>
            </div>
            <div>
              <label for="inventory-reason" class="block text-sm font-medium text-gray-700">{{ t('admin.inventory.reason') }}</label>
              <select id="inventory-reason" v-model="adjustmentReason" name="adjustmentReason" autocomplete="off" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500">
                <option value="restock">{{ t('admin.inventory.reason_restock') }}</option>
                <option value="correction">{{ t('admin.inventory.reason_correction') }}</option>
                <option value="damage">{{ t('admin.inventory.reason_damage') }}</option>
                <option value="return">{{ t('admin.inventory.reason_return') }}</option>
                <option value="audit">{{ t('admin.inventory.reason_audit') }}</option>
                <option value="other">{{ t('admin.inventory.reason_other') }}</option>
              </select>
            </div>
            <div>
              <label for="inventory-notes" class="block text-sm font-medium text-gray-700">{{ t('admin.inventory.notes') }}</label>
              <textarea id="inventory-notes" v-model="adjustmentNotes" name="adjustmentNotes" rows="2" autocomplete="off" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500"></textarea>
            </div>
            <div v-if="adjustmentError" class="text-sm text-red-600">{{ adjustmentError }}</div>
          </div>
          <div class="px-6 py-4 bg-gray-50 flex justify-end gap-3">
            <button @click="closeAdjustmentModal" class="px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100 transition-colors">
              {{ t('admin.inventory.cancel') }}
            </button>
            <button @click="submitAdjustment" :disabled="savingAdjustment" class="px-4 py-2 bg-orange-600 text-white rounded-lg text-sm font-medium hover:bg-orange-700 disabled:opacity-50 transition-colors flex items-center gap-2">
              <Icon v-if="savingAdjustment" name="heroicons:arrow-path" class="h-4 w-4 animate-spin" aria-hidden="true" />
              {{ t('admin.inventory.save') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- History Modal -->
    <div v-if="showHistoryModal" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-screen items-end justify-center px-4 pb-20 pt-4 text-center sm:block sm:p-0">
        <button type="button" class="fixed inset-0 w-full h-full bg-gray-500 bg-opacity-75 transition-opacity border-0 cursor-pointer" @click="showHistoryModal = false" :aria-label="t('close')" />
        <div class="inline-block w-full transform overflow-hidden rounded-xl bg-white text-start align-bottom shadow-xl sm:my-8 sm:max-w-2xl sm:align-middle">
          <div class="bg-gradient-to-r from-gray-700 to-gray-900 px-6 py-4 flex items-center justify-between">
            <div>
              <h3 class="text-lg font-semibold text-white">{{ t('admin.inventory.history_title') }}</h3>
              <p class="text-sm text-gray-300 mt-0.5">{{ tField(historyProduct, 'name') }}</p>
            </div>
            <button @click="showHistoryModal = false" class="text-gray-300 hover:text-white" :aria-label="t('close')">
              <Icon name="heroicons:x-mark" class="h-5 w-5" aria-hidden="true" />
            </button>
          </div>
          <div class="p-6 max-h-96 overflow-y-auto">
            <div v-if="stockHistory.length === 0" class="text-center py-8 text-gray-500">
              {{ t('admin.inventory.no_history') }}
            </div>
            <div v-else class="space-y-3">
              <div v-for="entry in stockHistory" :key="entry.id" class="flex items-start gap-4 p-3 rounded-lg border border-gray-100 hover:bg-gray-50">
                <div :class="entry.change > 0 ? 'bg-emerald-100 text-emerald-700' : 'bg-red-100 text-red-700'" class="flex-shrink-0 w-16 h-16 rounded-lg flex items-center justify-center font-bold text-lg">
                  {{ entry.change > 0 ? '+' : '' }}{{ entry.change }}
                </div>
                <div class="flex-1">
                  <div class="flex items-center justify-between">
                    <span class="font-medium text-gray-900">{{ entry.reason }}</span>
                    <span class="text-xs text-gray-500">{{ formatDate(entry.createdAt, { dateStyle: 'medium', timeStyle: 'short' }) }}</span>
                  </div>
                  <p class="text-sm text-gray-600 mt-0.5">{{ entry.notes }}</p>
                  <div class="text-xs text-gray-400 mt-1">
                    {{ t('admin.inventory.from_to', { from: entry.previousQty, to: entry.newQty }) }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Warehouse Modal -->
    <div v-if="showWarehouseModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <div class="bg-white rounded-xl shadow-xl w-full max-w-lg p-6">
        <h3 class="text-lg font-semibold mb-4">{{ warehouseForm.id ? t('admin.inventory.edit') : t('admin.inventory.add_warehouse') }}</h3>
        <form class="space-y-3" @submit.prevent="saveWarehouse">
          <input v-model="warehouseForm.id" :placeholder="t('admin.inventory.wh_id')" required class="w-full border rounded-lg px-3 py-2 text-sm font-mono" />
          <input v-model="warehouseForm.code" :placeholder="t('admin.inventory.wh_code')" required class="w-full border rounded-lg px-3 py-2 text-sm" />
          <input v-model="warehouseForm.name" :placeholder="t('admin.inventory.wh_name')" required class="w-full border rounded-lg px-3 py-2 text-sm" />
          <input v-model="warehouseForm.country" :placeholder="t('admin.inventory.wh_country')" class="w-full border rounded-lg px-3 py-2 text-sm" />
          <div class="flex justify-end gap-2 pt-2">
            <button type="button" class="px-4 py-2 border rounded-lg text-sm" @click="showWarehouseModal = false">{{ t('admin.inventory.cancel') }}</button>
            <button type="submit" class="px-4 py-2 bg-orange-600 text-white rounded-lg text-sm">{{ t('admin.inventory.save') }}</button>
          </div>
        </form>
      </div>
    </div>

    <!-- Warehouse Stock Modal -->
    <div v-if="showWarehouseStockModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <div class="bg-white rounded-xl shadow-xl w-full max-w-lg p-6">
        <h3 class="text-lg font-semibold mb-4">{{ t('admin.inventory.update_stock') }} — {{ stockWarehouse?.name }}</h3>
        <form class="space-y-3" @submit.prevent="saveWarehouseStock">
          <input v-model="warehouseStockForm.productId" :placeholder="t('admin.inventory.product_id')" required class="w-full border rounded-lg px-3 py-2 text-sm font-mono" />
          <input v-model.number="warehouseStockForm.quantity" type="number" min="0" :placeholder="t('admin.inventory.col_stock')" required class="w-full border rounded-lg px-3 py-2 text-sm" />
          <div class="flex justify-end gap-2 pt-2">
            <button type="button" class="px-4 py-2 border rounded-lg text-sm" @click="showWarehouseStockModal = false">{{ t('admin.inventory.cancel') }}</button>
            <button type="submit" class="px-4 py-2 bg-orange-600 text-white rounded-lg text-sm">{{ t('admin.inventory.save') }}</button>
          </div>
        </form>
      </div>
    </div>

    <!-- Transfer Modal -->
    <div v-if="showTransferModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <div class="bg-white rounded-xl shadow-xl w-full max-w-lg p-6">
        <h3 class="text-lg font-semibold mb-4">{{ t('admin.inventory.create_transfer') }}</h3>
        <form class="space-y-3" @submit.prevent="saveTransfer">
          <input v-model="transferForm.fromWarehouseId" :placeholder="t('admin.inventory.from_wh')" required class="w-full border rounded-lg px-3 py-2 text-sm" />
          <input v-model="transferForm.toWarehouseId" :placeholder="t('admin.inventory.to_wh')" required class="w-full border rounded-lg px-3 py-2 text-sm" />
          <input v-model="transferForm.productId" :placeholder="t('admin.inventory.product_id')" required class="w-full border rounded-lg px-3 py-2 text-sm font-mono" />
          <input v-model.number="transferForm.quantity" type="number" min="1" :placeholder="t('admin.inventory.qty')" required class="w-full border rounded-lg px-3 py-2 text-sm" />
          <textarea v-model="transferForm.notes" rows="2" :placeholder="t('admin.inventory.notes')" class="w-full border rounded-lg px-3 py-2 text-sm" />
          <div class="flex justify-end gap-2 pt-2">
            <button type="button" class="px-4 py-2 border rounded-lg text-sm" @click="showTransferModal = false">{{ t('admin.inventory.cancel') }}</button>
            <button type="submit" class="px-4 py-2 bg-orange-600 text-white rounded-lg text-sm">{{ t('admin.inventory.save') }}</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useTranslation } from '~/composables/useTranslation'

definePageMeta({ layout: 'admin', middleware: ['auth'] })

const api = useApi()
const { enableMultiWarehouse } = useFeatureFlags()
const { t } = useI18n()
const { tField } = useTranslation()
const { currencyOrDefault: cur, formatNumber, formatDate } = useDisplay()

const items = ref<any[]>([])
const categoryCatalog = ref<any[]>([])
const pagination = ref<any>(null)
const pending = ref(true)
const error = ref('')
const page = ref(1)
const pageSize = 20
const searchQuery = ref('')
const stockFilter = ref('all')
const categoryFilter = ref('')

// Filters are client-side (the /admin/inventory endpoint has no filter query
// params), so when any filter is active we load the WHOLE dataset and filter +
// paginate locally. Otherwise the filters would only ever apply to the current
// server page — search/category/stock results silently missing other pages.
const allInventory = ref<any[]>([])
let allInventoryFetched = false

const filterActive = computed(() =>
  !!searchQuery.value || stockFilter.value !== 'all' || !!categoryFilter.value,
)
const showAdjustmentModal = ref(false)
const adjustmentProduct = ref<any>(null)
const newStock = ref(0)
const adjustmentType = ref('set')
const adjustmentReason = ref('restock')
const adjustmentNotes = ref('')
const savingAdjustment = ref(false)
const adjustmentError = ref('')
const showHistoryModal = ref(false)
const historyProduct = ref<any>(null)
const stockHistory = ref<any[]>([])

// Warehouses & transfers tabs
const inventoryTab = ref<'stock' | 'warehouses' | 'transfers'>('stock')
const warehouses = ref<any[]>([])
const transfers = ref<any[]>([])
const warehousesPending = ref(false)
const transfersPending = ref(false)
const showWarehouseModal = ref(false)
const showWarehouseStockModal = ref(false)
const showTransferModal = ref(false)
const stockWarehouse = ref<any>(null)
const warehouseForm = reactive({ id: '', code: '', name: '', country: '', type: 'factory', isActive: true })
const warehouseStockForm = reactive({ productId: '', quantity: 0 })
const transferForm = reactive({ fromWarehouseId: '', toWarehouseId: '', productId: '', quantity: 1, notes: '' })

// Multi-select
const selectedIds = ref<Set<string>>(new Set())

// Import
const showImportModal = ref(false)
const importLoading = ref(false)
const importError = ref('')
const importPreview = ref<any>(null)
const importApplying = ref(false)
const importResult = ref<any>(null)
const fileInput = ref<HTMLInputElement>()

// Batch edit
const showBatchEditModal = ref(false)
const batchEditField = ref('')
const batchEditValue = ref<any>('')
const batchEditApplying = ref(false)
const batchEditError = ref('')
const batchEditResult = ref<any>(null)

// Batch delete
const showBatchDeleteModal = ref(false)
const batchDeleteApplying = ref(false)
const batchDeleteError = ref('')
const batchDeleteResult = ref<any>(null)

const stats = computed(() => {
  // When filters are active the counts must come from the full filtered set,
  // otherwise low/out/value only reflect the current page.
  const source = filterActive.value ? filteredFull.value : items.value
  const low = source.filter(i => i.stockQuantity > 0 && i.stockQuantity <= (i.moq || 10)).length
  const out = source.filter(i => !i.stockQuantity || i.stockQuantity <= 0).length
  const value = source.reduce((sum, i) => sum + ((i.stockQuantity || 0) * (i.basePrice || 0)), 0)
  return { totalProducts: displayPagination.value?.total || source.length, lowStock: low, outOfStock: out, totalValue: value, currency: '' }
})

const categoryOptions = computed(() =>
  categoryCatalog.value.map(cat => ({
    slug: cat.slug,
    label: tField(cat, 'name') || cat.name || cat.slug,
  })).sort((a, b) => a.label.localeCompare(b.label)),
)

const categoryBySlug = computed(() => {
  const map = new Map<string, any>()
  for (const cat of categoryCatalog.value) map.set(cat.slug, cat)
  return map
})

const categoryLabel = (item: any) => {
  const slug = item?.categorySlug
  if (slug && categoryBySlug.value.has(slug)) {
    const cat = categoryBySlug.value.get(slug)
    return tField(cat, 'name') || cat?.name || item.category || '-'
  }
  return item?.category || '-'
}

const matchesFilters = (item: any) => {
  const matchesSearch = !searchQuery.value || (item.name || '').toLowerCase().includes(searchQuery.value.toLowerCase()) || (item.slug || '').toLowerCase().includes(searchQuery.value.toLowerCase())
  const matchesCategory = !categoryFilter.value || item.categorySlug === categoryFilter.value
  const matchesStock = (stockFilter.value === 'all') || (stockFilter.value === 'low' && item.stockQuantity > 0 && item.stockQuantity <= (item.moq || 10)) || (stockFilter.value === 'out' && (!item.stockQuantity || item.stockQuantity <= 0)) || (stockFilter.value === 'in' && item.stockQuantity > (item.moq || 10))
  return matchesSearch && matchesCategory && matchesStock
}

/** Full filtered set: the whole dataset when filtering, the current page otherwise. */
const filteredFull = computed(() => {
  const source = filterActive.value ? allInventory.value : items.value
  return source.filter(matchesFilters)
})

/** Rendered rows: client-side page slice when filtering, raw current page otherwise. */
const filteredItems = computed(() => {
  const list = filteredFull.value
  if (!filterActive.value) return list
  const start = (page.value - 1) * pageSize
  return list.slice(start, start + pageSize)
})

/** Pagination the footer/stats read: local filtered pagination while filtering. */
const displayPagination = computed(() => {
  if (!filterActive.value) return pagination.value
  const total = filteredFull.value.length
  return { total, page: page.value, limit: pageSize, totalPages: Math.max(1, Math.ceil(total / pageSize)) }
})

const isAllSelected = computed(() => {
  return filteredItems.value.length > 0 && filteredItems.value.every(i => selectedIds.value.has(i.id))
})

const activeMapping = computed(() => {
  const importPrev = importPreview.value
  if (!importPrev) return {}
  // Merge heuristic + AI mapping (AI takes precedence)
  return { ...(importPrev.columnMapping || {}), ...(importPrev.aiMapping || {}) }
})

const aiMappingSummary = computed(() => {
  const ai = importPreview.value?.aiMapping
  if (!ai) return ''
  return Object.entries(ai).map(([field, colIdx]) => {
    const headerName = importPreview.value.headers[parseInt(colIdx as string)] || colIdx
    return `${field} → ${headerName}`
  }).join(', ')
})

const clearFilters = () => {
  searchQuery.value = ''
  stockFilter.value = 'all'
  categoryFilter.value = ''
}

const clearSelection = () => {
  selectedIds.value = new Set()
}

// ── Batch Edit ──

const batchEditFields: Record<string, { type: string; options?: string[] }> = {
  category: { type: 'text' },
  categorySlug: { type: 'text' },
  status: { type: 'select', options: ['active', 'inactive', 'draft'] },
  stockQuantity: { type: 'number' },
  basePrice: { type: 'number' },
  moq: { type: 'number' },
  leadTime: { type: 'text' },
  halalCertified: { type: 'boolean' },
  oemAvailable: { type: 'boolean' },
  hsCode: { type: 'text' },
  shelfLife: { type: 'text' },
  storage: { type: 'text' },
}

const openBatchEditModal = () => {
  batchEditField.value = ''
  batchEditValue.value = ''
  batchEditError.value = ''
  batchEditResult.value = null
  showBatchEditModal.value = true
}

const closeBatchEditModal = () => {
  showBatchEditModal.value = false
  batchEditApplying.value = false
}

const submitBatchEdit = async () => {
  if (!batchEditField.value || batchEditValue.value === '' || batchEditValue.value === null) return
  batchEditApplying.value = true
  batchEditError.value = ''
  batchEditResult.value = null

  const updates: Record<string, any> = { [batchEditField.value]: batchEditValue.value }

  try {
    const result = await api.batchUpdateInventory(Array.from(selectedIds.value), updates)
    batchEditResult.value = result
    if (result.updated > 0) {
      allInventoryFetched = false
      await fetchInventory()
      clearSelection()
    }
  } catch (err: any) {
    batchEditError.value = err?.message || t('errors.api.load_failed')
  } finally {
    batchEditApplying.value = false
  }
}

const openBatchDeleteModal = () => {
  batchDeleteError.value = ''
  batchDeleteResult.value = null
  showBatchDeleteModal.value = true
}

const closeBatchDeleteModal = () => {
  showBatchDeleteModal.value = false
  batchDeleteApplying.value = false
}

const submitBatchDelete = async () => {
  batchDeleteApplying.value = true
  batchDeleteError.value = ''
  batchDeleteResult.value = null

  try {
    const result = await api.batchDeleteInventory(Array.from(selectedIds.value))
    batchDeleteResult.value = result
    if (result.deleted > 0) {
      allInventoryFetched = false
      await fetchInventory()
      clearSelection()
    }
  } catch (err: any) {
    batchDeleteError.value = err?.message || t('errors.api.load_failed')
  } finally {
    batchDeleteApplying.value = false
  }
}

// ── API ──

const fetchCategories = async () => {
  try {
    categoryCatalog.value = await api.adminGetCategories()
  } catch {
    categoryCatalog.value = []
  }
}

const fetchAllInventory = async () => {
  if (allInventoryFetched) return
  const out: any[] = []
  let p = 1
  let totalPages = 1
  do {
    const res = await api.get<any>('/admin/inventory', { page: p, limit: 100 })
    const rows = res.data || []
    out.push(...rows)
    totalPages = res.pagination?.totalPages || 1
    if (rows.length < 100) break
    p++
  } while (p <= totalPages && p <= 50) // safety cap; the admin backend caps limit at 100
  allInventory.value = out
  allInventoryFetched = true
}

const fetchInventory = async () => {
  pending.value = true; error.value = ''
  try {
    if (filterActive.value) {
      await fetchAllInventory()
    } else {
      const res = await api.get<any>('/admin/inventory', { page: page.value, limit: pageSize })
      items.value = res.data || []
      pagination.value = res.pagination
      // Leaving filter mode: invalidate the cached full dataset so the next
      // filter session re-fetches fresh data.
      allInventoryFetched = false
    }
  } catch (err: any) {
    error.value = err?.message || t('errors.api.load_failed')
  } finally { pending.value = false }
}

// ── Selection ──

const toggleSelect = (id: string) => {
  const next = new Set(selectedIds.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  selectedIds.value = next
}

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedIds.value = new Set()
  } else {
    selectedIds.value = new Set(filteredItems.value.map(i => i.id))
  }
}

// ── Export ──

const exportSelected = async () => {
  const ids = selectedIds.value.size > 0 ? Array.from(selectedIds.value) : filteredItems.value.map(i => i.id)
  if (ids.length === 0) return
  try {
    const blob = await api.exportInventoryXlsx(ids)
    const url = URL.createObjectURL(blob as any)
    const a = document.createElement('a')
    a.href = url
    a.download = `inventory_${new Date().toISOString().split('T')[0]}.xlsx`
    a.click()
    URL.revokeObjectURL(url)
  } catch (err: any) {
    console.error('Export failed:', err)
  }
}

// ── XLSX template (minimal OOXML) ──
// The importer (/admin/inventory/import-xlsx) only accepts .xlsx and validates
// the extension before excelize opens it, so the download template MUST be a
// real XLSX, not CSV. No spreadsheet library is bundled, so we build a minimal
// but valid .xlsx: a stored (uncompressed) ZIP containing the OOXML parts.
function crc32(buf: Uint8Array): number {
  let c = ~0
  for (let i = 0; i < buf.length; i++) {
    c ^= buf[i]
    for (let k = 0; k < 8; k++) c = c & 1 ? (c >>> 1) ^ 0xedb88320 : c >>> 1
  }
  return ~c >>> 0
}

function buildXlsxZip(files: { name: string; data: Uint8Array }[]): Uint8Array<ArrayBuffer> {
  const encoder = new TextEncoder()
  const chunks: Uint8Array[] = []
  const central: Uint8Array[] = []
  let offset = 0
  for (const f of files) {
    const name = encoder.encode(f.name)
    const crc = crc32(f.data)
    const local = new DataView(new ArrayBuffer(30))
    local.setUint32(0, 0x04034b50, true) // PK\x03\x04
    local.setUint16(4, 20, true)         // version needed to extract
    local.setUint16(6, 0, true)          // flags
    local.setUint16(8, 0, true)          // compression method: stored
    local.setUint16(10, 0, true)         // mod time
    local.setUint16(12, 0x21, true)      // mod date
    local.setUint32(14, crc, true)
    local.setUint32(18, f.data.length, true) // compressed size
    local.setUint32(22, f.data.length, true) // uncompressed size
    local.setUint16(26, name.length, true)
    local.setUint16(28, 0, true) // extra field length
    chunks.push(new Uint8Array(local.buffer), name, f.data)

    const cd = new DataView(new ArrayBuffer(46))
    cd.setUint32(0, 0x02014b50, true) // PK\x01\x02
    cd.setUint16(4, 20, true)         // version made by
    cd.setUint16(6, 20, true)         // version needed
    cd.setUint16(8, 0, true)          // flags
    cd.setUint16(10, 0, true)         // method: stored
    cd.setUint16(12, 0, true)         // mod time
    cd.setUint16(14, 0x21, true)      // mod date
    cd.setUint32(16, crc, true)
    cd.setUint32(20, f.data.length, true)
    cd.setUint32(24, f.data.length, true)
    cd.setUint16(28, name.length, true)
    cd.setUint16(30, 0, true) // extra
    cd.setUint16(32, 0, true) // comment
    cd.setUint16(34, 0, true) // disk start
    cd.setUint16(36, 0, true) // internal attrs
    cd.setUint32(38, 0, true) // external attrs
    cd.setUint32(42, offset, true) // local header offset
    central.push(new Uint8Array(cd.buffer), name)
    offset += 30 + name.length + f.data.length
  }

  const cdSize = central.reduce((s, c) => s + c.length, 0)
  const eocd = new DataView(new ArrayBuffer(22))
  eocd.setUint32(0, 0x06054b50, true) // PK\x05\x06
  eocd.setUint16(4, 0, true)
  eocd.setUint16(6, 0, true)
  eocd.setUint16(8, files.length, true)
  eocd.setUint16(10, files.length, true)
  eocd.setUint32(12, cdSize, true)
  eocd.setUint32(16, offset, true)
  eocd.setUint16(20, 0, true) // comment length

  const all = [...chunks, ...central, new Uint8Array(eocd.buffer)]
  const total = all.reduce((s, c) => s + c.length, 0)
  const out = new Uint8Array(total)
  let p = 0
  for (const c of all) { out.set(c, p); p += c.length }
  return out
}

function xlsxColName(index: number): string {
  let n = index
  let s = ''
  while (n >= 0) { s = String.fromCharCode(65 + (n % 26)) + s; n = Math.floor(n / 26) - 1 }
  return s
}

function buildInventoryTemplateXlsx(rows: string[][]): Blob {
  const esc = (v: string) => v.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
  const sheetData = rows.map((row, r) =>
    `<row r="${r + 1}">` +
    row.map((cell, c) =>
      `<c r="${xlsxColName(c)}${r + 1}" t="inlineStr"><is><t>${esc(cell)}</t></is></c>`,
    ).join('') +
    '</row>',
  ).join('')

  const parts: { name: string; data: Uint8Array }[] = [
    {
      name: '[Content_Types].xml',
      data: new TextEncoder().encode(
        '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>' +
        '<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">' +
        '<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>' +
        '<Default Extension="xml" ContentType="application/xml"/>' +
        '<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>' +
        '<Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>' +
        '</Types>',
      ),
    },
    {
      name: '_rels/.rels',
      data: new TextEncoder().encode(
        '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>' +
        '<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">' +
        '<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>' +
        '</Relationships>',
      ),
    },
    {
      name: 'xl/workbook.xml',
      data: new TextEncoder().encode(
        '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>' +
        '<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">' +
        '<sheets><sheet name="Inventory" sheetId="1" r:id="rId1"/></sheets>' +
        '</workbook>',
      ),
    },
    {
      name: 'xl/_rels/workbook.xml.rels',
      data: new TextEncoder().encode(
        '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>' +
        '<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">' +
        '<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>' +
        '</Relationships>',
      ),
    },
    {
      name: 'xl/worksheets/sheet1.xml',
      data: new TextEncoder().encode(
        '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>' +
        '<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>' +
        sheetData +
        '</sheetData></worksheet>',
      ),
    },
  ]
  const zip = buildXlsxZip(parts)
  return new Blob([zip], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' })
}

const downloadTemplate = async () => {
  // Generate a template XLSX (matching the importer's .xlsx-only contract)
  // with headers matching product fields.
  const headers = ['Name', 'Category', 'Category Slug', 'Base Price', 'MOQ', 'Stock', 'Lead Time', 'Halal', 'OEM', 'HS Code', 'Shelf Life', 'Storage', 'Status', 'Thumbnail', 'Images', 'Flavors', 'Shapes', 'Ingredients', 'Allergens', 'Certifications', 'Description', 'Summary']
  const exampleRow = ['Example Candy', 'Hard Candy', 'hard-candy', '0.50', '1000', '5000', '', 'Yes', 'Yes', '170490', '12 months', 'Cool dry place', 'active', 'https://example.com/image.jpg', 'https://example.com/img1.jpg, https://example.com/img2.jpg', 'Strawberry, Mint', 'Round, Star', 'Sugar, Glucose', 'None', 'ISO, HACCP', 'A delicious candy product', 'Premium quality candy']
  const blob = buildInventoryTemplateXlsx([headers, exampleRow])
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'inventory_import_template.xlsx'
  a.click()
  URL.revokeObjectURL(url)
}

// ── Import ──

const openImportModal = () => {
  showImportModal.value = true
  importPreview.value = null
  importError.value = ''
  importLoading.value = false
  importResult.value = null
}

const closeImportModal = () => {
  showImportModal.value = false
  resetImport()
}

const resetImport = () => {
  importPreview.value = null
  importError.value = ''
  importLoading.value = false
  importResult.value = null
  if (fileInput.value) fileInput.value.value = ''
}

const triggerFileInput = () => {
  fileInput.value?.click()
}

const handleFileDrop = (e: DragEvent) => {
  const file = e.dataTransfer?.files?.[0]
  if (file) processFile(file)
}

const handleFileSelect = (e: Event) => {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (file) processFile(file)
}

const processFile = async (file: File) => {
  if (!file.name.endsWith('.xlsx')) {
    importError.value = t('admin.inventory.import_select_file')
    return
  }
  importError.value = ''
  importLoading.value = true
  importResult.value = null
  try {
    const result = await api.importInventoryXlsx(file)
    importPreview.value = result
  } catch (err: any) {
    importError.value = err?.message || t('errors.api.load_failed')
  } finally {
    importLoading.value = false
  }
}

const applyImport = async () => {
  if (!importPreview.value) return
  importApplying.value = true
  importResult.value = null
  try {
    // Build rows using column mapping
    const mapping = activeMapping.value
    const imageColIdx = importPreview.value.imageColumns || []
    const imageColNames = imageColIdx.map((i: number) => importPreview.value.headers[i])

    const rows = importPreview.value.rows.map((row: string[]) => {
      const mapped: Record<string, any> = {}
      // Apply mapped fields
      for (const [field, colIdx] of Object.entries(mapping)) {
        const idx = parseInt(colIdx as string)
        if (idx >= 0 && idx < row.length) {
          mapped[field] = row[idx]
        }
      }
      // Also include raw column data by header name for unmapped fields
      importPreview.value.headers.forEach((h: string, i: number) => {
        if (!(h in mapped) && i < row.length) {
          mapped[h] = row[i]
        }
      })
      return mapped
    })

    const result = await api.applyInventoryImport({ rows, imageColumns: imageColNames })
    importResult.value = result
    if (result.errors === 0 || result.created > 0 || result.updated > 0) {
      allInventoryFetched = false
      await fetchInventory()
    }
  } catch (err: any) {
    importResult.value = { errors: 1, created: 0, updated: 0, errorMessages: [err?.message || 'Import failed'] }
  } finally {
    importApplying.value = false
  }
}

const isImageURL = (str: string) => {
  if (!str) return false
  return /^https?:\/\/.*\.(jpg|jpeg|png|gif|webp|svg|bmp)(\?.*)?$/i.test(str)
}

const hideBrokenImage = (event: Event) => {
  const img = event.target as HTMLImageElement | null
  if (img) img.style.display = 'none'
}

// ── Stock Adjustment ──

const openAdjustmentModal = (product?: any) => {
  adjustmentProduct.value = product || null; newStock.value = product?.stockQuantity || 0
  adjustmentType.value = 'set'; adjustmentReason.value = 'restock'; adjustmentNotes.value = ''; adjustmentError.value = ''; showAdjustmentModal.value = true
}
const closeAdjustmentModal = () => { showAdjustmentModal.value = false; savingAdjustment.value = false }

const submitAdjustment = async () => {
  if (!adjustmentProduct.value) return
  savingAdjustment.value = true; adjustmentError.value = ''
  let finalQty = newStock.value
  if (adjustmentType.value === 'add') finalQty = (adjustmentProduct.value.stockQuantity || 0) + newStock.value
  else if (adjustmentType.value === 'subtract') finalQty = Math.max(0, (adjustmentProduct.value.stockQuantity || 0) - newStock.value)
  try {
    await api.put(`/admin/inventory/${adjustmentProduct.value.id}`, { stockQuantity: finalQty, reason: adjustmentReason.value, notes: adjustmentNotes.value })
    allInventoryFetched = false
    closeAdjustmentModal(); await fetchInventory()
  } catch (err: any) { adjustmentError.value = err?.message || t('errors.api.stock_adjust_failed') }
  finally { savingAdjustment.value = false }
}

const viewHistory = async (product: any) => {
  historyProduct.value = product; showHistoryModal.value = true
  try { stockHistory.value = await api.get<any[]>(`/admin/inventory/${product.id}/history`) || [] }
  catch { stockHistory.value = [] }
}

const stockStatusClass = (qty: number, moq: number) => { if (!qty || qty <= 0) return 'text-red-600'; if (qty <= (moq || 10)) return 'text-orange-600'; return 'text-emerald-600' }
const stockBadgeClass = (qty: number, moq: number) => { if (!qty || qty <= 0) return 'bg-red-100 text-red-800'; if (qty <= (moq || 10)) return 'bg-orange-100 text-orange-800'; return 'bg-emerald-100 text-emerald-800' }
const stockStatusLabel = (qty: number, moq: number) => { if (!qty || qty <= 0) return t('admin.inventory.out_of_stock'); if (qty <= (moq || 10)) return t('admin.inventory.low_stock'); return t('admin.inventory.in_stock') }
const prevPage = () => { if (page.value > 1) { page.value -= 1; fetchInventory() } }
const nextPage = () => { if (displayPagination.value && page.value < displayPagination.value.totalPages) { page.value += 1; fetchInventory() } }

// Filters run client-side over the full dataset; re-slice from page 1 whenever
// a filter changes so results stay page-correct.
watch([searchQuery, stockFilter, categoryFilter], () => {
  page.value = 1
  fetchInventory()
})

const fetchWarehouses = async () => {
  warehousesPending.value = true
  try { warehouses.value = (await api.get<any[]>('/admin/warehouses')) || [] }
  catch { warehouses.value = [] }
  finally { warehousesPending.value = false }
}

const fetchTransfers = async () => {
  transfersPending.value = true
  try {
    const res = await api.get<any>('/admin/inventory/transfers', { page: 1, limit: 50 })
    transfers.value = res.data || []
  } catch { transfers.value = [] }
  finally { transfersPending.value = false }
}

const openWarehouseModal = (wh?: any) => {
  if (wh) Object.assign(warehouseForm, { id: wh.id, code: wh.code, name: wh.name, country: wh.country || '', type: wh.type || 'factory', isActive: wh.isActive !== false })
  else Object.assign(warehouseForm, { id: '', code: '', name: '', country: '', type: 'factory', isActive: true })
  showWarehouseModal.value = true
}

const saveWarehouse = async () => {
  await api.post('/admin/warehouses', { ...warehouseForm })
  showWarehouseModal.value = false
  await fetchWarehouses()
}

const openWarehouseStockModal = (wh: any) => {
  stockWarehouse.value = wh
  warehouseStockForm.productId = ''
  warehouseStockForm.quantity = 0
  showWarehouseStockModal.value = true
}

const saveWarehouseStock = async () => {
  await api.put(`/admin/warehouses/${stockWarehouse.value.id}/stock`, { ...warehouseStockForm })
  showWarehouseStockModal.value = false
}

const openTransferModal = () => {
  Object.assign(transferForm, { fromWarehouseId: '', toWarehouseId: '', productId: '', quantity: 1, notes: '' })
  showTransferModal.value = true
}

const saveTransfer = async () => {
  await api.post('/admin/inventory/transfer', {
    fromWarehouseId: transferForm.fromWarehouseId,
    toWarehouseId: transferForm.toWarehouseId,
    notes: transferForm.notes,
    items: [{ productId: transferForm.productId, quantity: transferForm.quantity }],
  })
  showTransferModal.value = false
  await fetchTransfers()
}

onMounted(async () => {
  await fetchCategories()
  await fetchInventory()
})
</script>
