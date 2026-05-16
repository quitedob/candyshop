<template>
  <div>
    <div class="sm:flex sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900">{{ t('admin.products.title') }}</h1>
        <p class="mt-2 text-sm text-gray-700">{{ t('admin.products.description') }}</p>
      </div>
      <div class="mt-4 sm:mt-0">
        <button type="button" class="inline-flex items-center justify-center rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-orange-700" @click="openCreateModal">
          {{ t('admin.products.add_product') }}
        </button>
      </div>
    </div>

    <div class="mt-8 overflow-hidden rounded-lg bg-white shadow ring-1 ring-black ring-opacity-5">
      <table class="min-w-full divide-y divide-gray-300">
        <thead class="bg-gray-50">
          <tr>
            <th class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6">{{ t('admin.products.col_product') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.products.col_category') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.products.col_moq') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.products.col_stock') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.products.col_status') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ $t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200 bg-white">
          <tr v-if="pending">
            <td colspan="6" class="py-5 text-center text-sm text-gray-500">{{ t('admin.products.loading') }}</td>
          </tr>
          <tr v-else-if="error">
            <td colspan="6" class="py-5 text-center text-sm text-red-600">{{ error }}</td>
          </tr>
          <tr v-else-if="products.length === 0">
            <td colspan="6" class="py-5 text-center text-sm text-gray-500">{{ t('admin.products.no_data') }}</td>
          </tr>
          <tr v-else v-for="product in products" :key="product.id">
            <td class="py-4 pl-4 pr-3 text-sm sm:pl-6">
              <div class="flex items-center">
                <div class="h-10 w-10 flex-shrink-0">
                  <img v-if="product.thumbnail" class="h-10 w-10 rounded-full object-cover" :src="product.thumbnail" alt="" />
                  <div v-else class="flex h-10 w-10 items-center justify-center rounded-full bg-gray-200 font-bold text-gray-500">
                    {{ product.name?.charAt(0) }}
                  </div>
                </div>
                <div class="ml-4">
                  <div class="font-medium text-gray-900">{{ product.name }}</div>
                  <div class="text-gray-500">{{ product.slug }}</div>
                </div>
              </div>
            </td>
            <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{{ product.category || '-' }}</td>
            <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{{ product.moq || 0 }}</td>
            <td class="whitespace-nowrap px-3 py-4 text-sm" :class="(product.stockQuantity || 0) > 0 ? 'text-light' : 'text-error'">
              {{ product.stockQuantity || 0 }}
            </td>
            <td class="whitespace-nowrap px-3 py-4 text-sm text-light">
              <span class="badge" :class="product.status === 'active' ? 'badge-success' : 'badge-default'">
                {{ enumLabel('product_status', product.status, 'active') }}
              </span>
            </td>
            <td class="whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
              <button type="button" class="text-orange-600 hover:text-orange-900" @click="openEditModal(product)">{{ t('admin.products.edit') }}</button>
              <button type="button" class="ml-4 text-red-600 hover:text-red-900" @click="deleteProduct(product.id)">{{ t('admin.products.delete') }}</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="pagination" class="mt-4 flex items-center justify-between rounded-lg border-t border-gray-200 bg-white px-4 py-3 shadow sm:px-6">
      <div class="text-sm text-gray-700">
        {{ t('admin.products.showing', { from: ((page - 1) * pageSize) + 1, to: Math.min(page * pageSize, pagination.total), total: pagination.total }) }}
      </div>
      <div class="flex items-center gap-2">
        <button type="button" class="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-700 disabled:opacity-50" :disabled="page <= 1" @click="prevPage">
          {{ t('admin.products.previous') }}
        </button>
        <button type="button" class="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-700 disabled:opacity-50" :disabled="page >= pagination.totalPages" @click="nextPage">
          {{ t('admin.products.next') }}
        </button>
      </div>
    </div>

    <div v-if="showModal" class="fixed inset-0 z-10 overflow-y-auto" role="dialog" aria-modal="true">
      <div class="flex min-h-screen items-end justify-center px-4 pb-20 pt-4 text-center sm:block sm:p-0">
        <button type="button" class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity w-full border-0 cursor-pointer" @click="closeModal" :aria-label="t('close')"></button>
        <span class="hidden sm:inline-block sm:h-screen sm:align-middle" aria-hidden="true">&#8203;</span>
        <div class="inline-block w-full transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left align-bottom shadow-xl sm:my-8 sm:max-w-3xl sm:p-6 sm:align-middle">
          <h3 class="text-lg font-medium leading-6 text-gray-900">{{ editingId ? t('admin.products.edit_product') : t('admin.products.create_product') }}</h3>

          <form class="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2" @submit.prevent="saveProduct">
            <div>
              <label for="product-name" class="block text-sm font-medium text-gray-700">{{ t('admin.products.name') }}</label>
              <input id="product-name" v-model="form.name" name="name" autocomplete="off" required class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-slug" class="block text-sm font-medium text-gray-700">{{ t('admin.products.slug') }}</label>
              <input id="product-slug" v-model="form.slug" name="slug" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-category" class="block text-sm font-medium text-gray-700">{{ t('admin.products.category') }}</label>
              <input id="product-category" v-model="form.category" name="category" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-categorySlug" class="block text-sm font-medium text-gray-700">{{ t('admin.products.category_slug') }}</label>
              <input id="product-categorySlug" v-model="form.categorySlug" name="categorySlug" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-moq" class="block text-sm font-medium text-gray-700">{{ t('admin.products.moq') }}</label>
              <input id="product-moq" v-model.number="form.moq" name="moq" type="number" min="1" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-basePrice" class="block text-sm font-medium text-gray-700">{{ t('admin.products.base_price') }}</label>
              <input id="product-basePrice" v-model.number="form.basePrice" name="basePrice" type="number" min="0.01" step="0.01" required autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-leadTime" class="block text-sm font-medium text-gray-700">{{ t('admin.products.lead_time') }}</label>
              <input id="product-leadTime" v-model="form.leadTime" name="leadTime" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-stockQuantity" class="block text-sm font-medium text-gray-700">{{ t('admin.products.stock_quantity') }}</label>
              <input id="product-stockQuantity" v-model.number="form.stockQuantity" name="stockQuantity" type="number" min="0" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-status" class="block text-sm font-medium text-gray-700">{{ t('admin.products.status') }}</label>
              <select id="product-status" v-model="form.status" name="status" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
                <option value="active">{{ enumLabel('product_status', 'active') }}</option>
                <option value="draft">{{ enumLabel('product_status', 'draft') }}</option>
                <option value="inactive">{{ enumLabel('product_status', 'inactive') }}</option>
              </select>
            </div>
            <div class="flex items-center gap-4 pt-7">
              <label class="inline-flex items-center">
                <input id="product-oemAvailable" v-model="form.oemAvailable" name="oemAvailable" type="checkbox" class="rounded border-gray-300" />
                <span class="ml-2 text-sm text-gray-700">{{ t('admin.products.oem') }}</span>
              </label>
              <label class="inline-flex items-center">
                <input id="product-halalCertified" v-model="form.halalCertified" name="halalCertified" type="checkbox" class="rounded border-gray-300" />
                <span class="ml-2 text-sm text-gray-700">{{ t('admin.products.halal') }}</span>
              </label>
              <label class="inline-flex items-center">
                <input id="product-featured" v-model="form.featured" name="featured" type="checkbox" class="rounded border-gray-300" />
                <span class="ml-2 text-sm text-gray-700">{{ t('admin.products.featured') }}</span>
              </label>
            </div>

            <!-- Weight & Measurement -->
            <div class="sm:col-span-2 mt-2 border-t pt-3">
              <h4 class="text-sm font-semibold text-gray-900 mb-2">{{ t('admin.products.section_weight') }}</h4>
            </div>
            <div>
              <label for="product-netWeightPerPiece" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_net_weight_per_piece') }}</label>
              <input id="product-netWeightPerPiece" v-model.number="form.netWeightPerPiece" type="number" step="0.1" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-netWeightPerPack" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_net_weight_per_pack') }}</label>
              <input id="product-netWeightPerPack" v-model.number="form.netWeightPerPack" type="number" step="0.1" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-grossWeightPerCarton" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_gross_weight_per_carton') }}</label>
              <input id="product-grossWeightPerCarton" v-model.number="form.grossWeightPerCarton" type="number" step="0.1" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-piecesPerPack" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_pieces_per_pack') }}</label>
              <input id="product-piecesPerPack" v-model.number="form.piecesPerPack" type="number" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-packsPerCarton" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_packs_per_carton') }}</label>
              <input id="product-packsPerCarton" v-model.number="form.packsPerCarton" type="number" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>

            <!-- Dimensions -->
            <div class="sm:col-span-2 mt-2 border-t pt-3">
              <h4 class="text-sm font-semibold text-gray-900 mb-2">{{ t('admin.products.section_dimensions') }}</h4>
            </div>
            <div>
              <label for="product-productLengthMM" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_length_mm') }}</label>
              <input id="product-productLengthMM" v-model.number="form.productLengthMM" type="number" step="0.1" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-productWidthMM" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_width_mm') }}</label>
              <input id="product-productWidthMM" v-model.number="form.productWidthMM" type="number" step="0.1" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-productHeightMM" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_height_mm') }}</label>
              <input id="product-productHeightMM" v-model.number="form.productHeightMM" type="number" step="0.1" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>

            <!-- Dietary -->
            <div class="sm:col-span-2 mt-2 border-t pt-3">
              <h4 class="text-sm font-semibold text-gray-900 mb-2">{{ t('admin.products.section_dietary') }}</h4>
            </div>
            <div class="sm:col-span-2 flex items-center gap-4">
              <label class="inline-flex items-center">
                <input v-model="form.isVegan" type="checkbox" class="rounded border-gray-300" />
                <span class="ml-2 text-sm text-gray-700">{{ t('admin.products.field_vegan') }}</span>
              </label>
              <label class="inline-flex items-center">
                <input v-model="form.isGlutenFree" type="checkbox" class="rounded border-gray-300" />
                <span class="ml-2 text-sm text-gray-700">{{ t('admin.products.field_gluten_free') }}</span>
              </label>
              <label class="inline-flex items-center">
                <input v-model="form.isSugarFree" type="checkbox" class="rounded border-gray-300" />
                <span class="ml-2 text-sm text-gray-700">{{ t('admin.products.field_sugar_free') }}</span>
              </label>
              <label class="inline-flex items-center">
                <input v-model="form.isKosher" type="checkbox" class="rounded border-gray-300" />
                <span class="ml-2 text-sm text-gray-700">{{ t('admin.products.field_kosher') }}</span>
              </label>
              <label class="inline-flex items-center">
                <input v-model="form.isOrganic" type="checkbox" class="rounded border-gray-300" />
                <span class="ml-2 text-sm text-gray-700">{{ t('admin.products.field_organic') }}</span>
              </label>
            </div>
            <div class="sm:col-span-2">
              <label for="product-summary" class="block text-sm font-medium text-gray-700">{{ t('admin.products.summary') }}</label>
              <textarea id="product-summary" v-model="form.summary" name="summary" rows="2" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea>
            </div>
            <div class="sm:col-span-2">
              <label for="product-description" class="block text-sm font-medium text-gray-700">{{ t('admin.products.description_label') }}</label>
              <textarea id="product-description" v-model="form.description" name="description" rows="3" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea>
            </div>
            <div class="sm:col-span-2">
              <label for="product-thumbnail" class="block text-sm font-medium text-gray-700">{{ t('admin.products.thumbnail') }}</label>
              <div class="mt-1 flex items-center gap-2">
                <input id="product-thumbnail" v-model="form.thumbnail" name="thumbnail" autocomplete="off" class="block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
                <button type="button" class="rounded-md bg-orange-50 px-3 py-2 text-xs font-medium text-orange-600 hover:bg-orange-500" @click="triggerUpload('thumbnail')">
                  {{ t('admin.products.upload_image') }}
                </button>
              </div>
            </div>
            <div class="sm:col-span-2">
              <label for="product-images" class="block text-sm font-medium text-gray-700">{{ t('admin.products.images') }}</label>
              <div class="mt-1 flex items-center gap-2">
                <input id="product-images" v-model="form.imagesInput" name="images" autocomplete="off" class="block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
                <button type="button" class="rounded-md bg-orange-50 px-3 py-2 text-xs font-medium text-orange-600 hover:text-orange-500" @click="triggerUpload('images')">
                  {{ t('admin.products.upload_images') }}
                </button>
                <div v-if="uploadingImage" class="mt-1 text-xs text-gray-500">
                  <button type="button" class="rounded-md bg-red-50 px-3 py-2 text-xs font-medium text-red-600 hover:text-red-500" @click="cancelImageUpload">
                    {{ t('admin.products.cancel') }}
                  </button>
                </div>
              </div>
            </div>
            <div>
              <label for="product-flavors" class="block text-sm font-medium text-gray-700">{{ t('admin.products.flavors') }}</label>
              <input id="product-flavors" v-model="form.flavorsInput" name="flavors" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-shapes" class="block text-sm font-medium text-gray-700">{{ t('admin.products.shapes') }}</label>
              <input id="product-shapes" v-model="form.shapesInput" name="shapes" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div class="sm:col-span-2">
              <label for="product-certifications" class="block text-sm font-medium text-gray-700">{{ t('admin.products.certifications') }}</label>
              <input id="product-certifications" v-model="form.certificationsInput" name="certifications" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-ingredients" class="block text-sm font-medium text-gray-700">{{ t('admin.products.ingredients') }}</label>
              <textarea id="product-ingredients" v-model="form.ingredients" name="ingredients" rows="2" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea>
            </div>
            <div>
              <label for="product-allergens" class="block text-sm font-medium text-gray-700">{{ t('admin.products.allergens') }}</label>
              <textarea id="product-allergens" v-model="form.allergens" name="allergens" rows="2" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea>
            </div>
            <div>
              <label for="product-shelfLife" class="block text-sm font-medium text-gray-700">{{ t('admin.products.shelf_life') }}</label>
              <input id="product-shelfLife" v-model="form.shelfLife" name="shelfLife" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-storage" class="block text-sm font-medium text-gray-700">{{ t('admin.products.storage') }}</label>
              <input id="product-storage" v-model="form.storage" name="storage" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>

            <!-- Nutrition per 100g -->
            <div class="sm:col-span-2 mt-2 border-t pt-3">
              <h4 class="text-sm font-semibold text-gray-900 mb-2">{{ t('admin.products.section_nutrition') }}</h4>
            </div>
            <div>
              <label for="product-energyKj" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_energy_kj') }}</label>
              <input id="product-energyKj" v-model.number="form.energyKj" type="number" step="1" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-energyKcal" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_energy_kcal') }}</label>
              <input id="product-energyKcal" v-model.number="form.energyKcal" type="number" step="1" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-totalFatG" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_fat_g') }}</label>
              <input id="product-totalFatG" v-model.number="form.totalFatG" type="number" step="0.1" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-saturatedFatG" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_saturated_fat_g') }}</label>
              <input id="product-saturatedFatG" v-model.number="form.saturatedFatG" type="number" step="0.1" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-carbohydratesG" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_carbohydrates_g') }}</label>
              <input id="product-carbohydratesG" v-model.number="form.carbohydratesG" type="number" step="0.1" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-sugarsG" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_sugars_g') }}</label>
              <input id="product-sugarsG" v-model.number="form.sugarsG" type="number" step="0.1" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-proteinG" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_protein_g') }}</label>
              <input id="product-proteinG" v-model.number="form.proteinG" type="number" step="0.1" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-saltG" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_salt_g') }}</label>
              <input id="product-saltG" v-model.number="form.saltG" type="number" step="0.01" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-fiberG" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_fiber_g') }}</label>
              <input id="product-fiberG" v-model.number="form.fiberG" type="number" step="0.1" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>

            <!-- Ingredient Compliance -->
            <div class="sm:col-span-2 mt-2 border-t pt-3">
              <h4 class="text-sm font-semibold text-gray-900 mb-2">{{ t('admin.products.section_compliance') }}</h4>
            </div>
            <div>
              <label for="product-additives" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_additives') }}</label>
              <input id="product-additives" v-model="form.additivesInput" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-sweetenerType" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_sweetener_type') }}</label>
              <input id="product-sweetenerType" v-model="form.sweetenerType" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-cocoaSolidsPct" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_cocoa_solids_pct') }}</label>
              <input id="product-cocoaSolidsPct" v-model.number="form.cocoaSolidsPct" type="number" step="0.1" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-milkSolidsPct" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_milk_solids_pct') }}</label>
              <input id="product-milkSolidsPct" v-model.number="form.milkSolidsPct" type="number" step="0.1" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-gmoStatus" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_gmo_status') }}</label>
              <select id="product-gmoStatus" v-model="form.gmoStatus" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
                <option value="">—</option>
                <option value="GMO">GMO</option>
                <option value="Non-GMO">Non-GMO</option>
                <option value="GMO-Free Certified">GMO-Free Certified</option>
              </select>
            </div>
            <div>
              <label for="product-mayContain" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_may_contain') }}</label>
              <input id="product-mayContain" v-model="form.mayContainInput" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-waterActivity" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_water_activity') }}</label>
              <input id="product-waterActivity" v-model.number="form.waterActivity" type="number" step="0.01" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>

            <!-- Trade & Packaging -->
            <div class="sm:col-span-2 mt-2 border-t pt-3">
              <h4 class="text-sm font-semibold text-gray-900 mb-2">{{ t('admin.products.section_trade') }}</h4>
            </div>
            <div>
              <label for="product-gtin" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_gtin') }}</label>
              <input id="product-gtin" v-model="form.gtin" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-hsCode" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_hs_code') }}</label>
              <input id="product-hsCode" v-model="form.hsCode" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-primaryPackaging" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_primary_packaging') }}</label>
              <input id="product-primaryPackaging" v-model="form.primaryPackaging" autocomplete="off" placeholder="flow-wrap, foil, box, bag, jar" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-innerPackConfig" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_inner_pack_config') }}</label>
              <input id="product-innerPackConfig" v-model="form.innerPackConfig" autocomplete="off" placeholder="e.g. 12 units per display box" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-palletConfig" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_pallet_config') }}</label>
              <input id="product-palletConfig" v-model="form.palletConfig" autocomplete="off" placeholder="e.g. 48 cases/layer × 5 layers" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>

            <!-- Sample Specs -->
            <div class="sm:col-span-2 mt-2 border-t pt-3">
              <h4 class="text-sm font-semibold text-gray-900 mb-2">{{ t('admin.products.section_sample') }}</h4>
            </div>
            <div>
              <label for="product-sampleMOQ" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_sample_moq') }}</label>
              <input id="product-sampleMOQ" v-model.number="form.sampleMOQ" type="number" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-sampleLeadTime" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_sample_lead_time') }}</label>
              <input id="product-sampleLeadTime" v-model="form.sampleLeadTime" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="product-samplePrice" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_sample_price') }}</label>
              <input id="product-samplePrice" v-model.number="form.samplePrice" type="number" step="0.01" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>

            <!-- Translations -->
            <div class="sm:col-span-2 mt-2 border-t pt-3">
              <div class="flex items-center justify-between mb-2">
                <h4 class="text-sm font-semibold text-gray-900">{{ t('admin.products.translations_label') }}</h4>
                <button
                  v-if="editingId"
                  type="button"
                  :disabled="aiTranslating"
                  class="inline-flex items-center gap-1.5 rounded-lg border border-orange-200 bg-orange-50 px-3 py-1 text-xs font-medium text-orange-700 hover:bg-orange-100 disabled:opacity-50 transition-colors"
                  @click="aiTranslateAll"
                >
                  <Icon name="heroicons:sparkles" class="h-3.5 w-3.5" />
                  {{ aiTranslating ? t('admin.products.ai_translating') : t('admin.products.ai_translate') }}
                </button>
              </div>
              <div class="flex gap-2 mb-3 flex-wrap">
                <button
                  v-for="loc in translationLocales"
                  :key="loc"
                  type="button"
                  :class="[
                    'rounded-lg px-3 py-1 text-xs font-medium transition-colors',
                    translationLocale === loc ? 'bg-orange-500 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                  ]"
                  @click="translationLocale = loc"
                >
                  {{ localeTabLabel(loc) }}
                </button>
              </div>
              <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div>
                  <label :for="`trans-name-${translationLocale}`" class="block text-sm font-medium text-gray-700">{{ t('admin.products.name') }} ({{ translationLocale }})</label>
                  <input :id="`trans-name-${translationLocale}`" v-model="form.translations[translationLocale].name" :name="`transName-${translationLocale}`" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
                </div>
                <div>
                  <label :for="`trans-summary-${translationLocale}`" class="block text-sm font-medium text-gray-700">{{ t('admin.products.summary') }} ({{ translationLocale }})</label>
                  <input :id="`trans-summary-${translationLocale}`" v-model="form.translations[translationLocale].summary" :name="`transSummary-${translationLocale}`" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
                </div>
                <div class="sm:col-span-2">
                  <label :for="`trans-desc-${translationLocale}`" class="block text-sm font-medium text-gray-700">{{ t('admin.products.description_label') }} ({{ translationLocale }})</label>
                  <textarea :id="`trans-desc-${translationLocale}`" v-model="form.translations[translationLocale].description" :name="`transDesc-${translationLocale}`" rows="2" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea>
                </div>
              </div>
            </div>

            <div v-if="formError" class="sm:col-span-2 text-sm text-red-600">{{ formError }}</div>
            <div class="sm:col-span-2 mt-2 flex justify-end gap-3">
              <button type="button" class="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm text-gray-700" @click="closeModal">
                {{ t('admin.products.cancel') }}
              </button>
              <button type="submit" :disabled="saving" class="rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm text-white disabled:opacity-50">
                {{ saving ? t('admin.products.saving') : (editingId ? t('admin.products.update') : t('admin.products.create')) }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <p v-if="actionMessage" class="mt-4 text-sm" :class="actionError ? 'text-red-600' : 'text-green-600'">
      {{ actionMessage }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, watch, onMounted } from 'vue'

definePageMeta({
  layout: 'admin',
  middleware: ['auth']
})

const api = useApi()
const { t } = useI18n()
const { enumLabel } = useDisplay()

const products = ref<any[]>([])
const pagination = ref<any>(null)
const pending = ref(true)
const error = ref('')
const page = ref(1)
const pageSize = 20

const showModal = ref(false)
const editingId = ref('')
const saving = ref(false)
const translationLocale = ref('en')
const aiTranslating = ref(false)
const formError = ref('')
const actionMessage = ref('')
const actionError = ref(false)
const uploadingImage = ref(false)
const uploadError = ref('')
let uploadAbortController: AbortController | null = null

const triggerUpload = async (field: 'thumbnail' | 'images') => {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = 'image/jpeg,image/png,image/webp,image/gif'
  input.onchange = async (e: any) => {
    const file = e.target.files?.[0]
    if (!file) return
    uploadingImage.value = true
    uploadError.value = ''
    uploadAbortController = new AbortController()
    try {
      const formData = new FormData()
      formData.append('file', file)
      const result = await api.post<any>('/admin/upload/image', formData)
      if (field === 'thumbnail') {
        form.thumbnail = result.url
      } else {
        const existing = form.imagesInput ? form.imagesInput.split(',').filter(Boolean) : []
        existing.push(result.url)
        form.imagesInput = existing.join(',')
      }
    } catch (err: any) {
      if (err?.name !== 'AbortError') {
        uploadError.value = err?.message || t('errors.api.upload_failed')
      }
    } finally {
      uploadingImage.value = false
      uploadAbortController = null
    }
  }
  input.click()
}

const cancelImageUpload = () => {
  if (uploadAbortController) {
    uploadAbortController.abort()
    uploadAbortController = null
  }
  uploadingImage.value = false
}

const translationLocales = ['en', 'zh', 'ko', 'ar', 'ja', 'th', 'vi', 'id', 'ms']

function emptyTranslations() {
  const result: Record<string, { name: string; summary: string; description: string }> = {}
  for (const loc of translationLocales) {
    result[loc] = { name: '', summary: '', description: '' }
  }
  return result
}

const form = reactive({
  name: '',
  slug: '',
  summary: '',
  description: '',
  category: '',
  categorySlug: '',
  thumbnail: '',
  moq: 0,
  basePrice: 0,
  stockQuantity: 0,
  leadTime: '',
  oemAvailable: false,
  halalCertified: false,
  featured: false,
  status: 'active',
  imagesInput: '',
  flavorsInput: '',
  shapesInput: '',
  certificationsInput: '',
  ingredients: '',
  allergens: '',
  shelfLife: '',
  storage: '',
  // Weight & Measurement
  netWeightPerPiece: 0,
  netWeightPerPack: 0,
  grossWeightPerCarton: 0,
  piecesPerPack: 0,
  packsPerCarton: 0,
  // Dimensions
  productLengthMM: 0,
  productWidthMM: 0,
  productHeightMM: 0,
  // Nutrition
  energyKj: 0,
  energyKcal: 0,
  totalFatG: 0,
  saturatedFatG: 0,
  carbohydratesG: 0,
  sugarsG: 0,
  proteinG: 0,
  saltG: 0,
  fiberG: 0,
  // Ingredient Compliance
  additivesInput: '',
  sweetenerType: '',
  cocoaSolidsPct: 0,
  milkSolidsPct: 0,
  gmoStatus: '',
  mayContainInput: '',
  waterActivity: 0,
  // Trade & Barcode
  gtin: '',
  hsCode: '',
  // Packaging
  primaryPackaging: '',
  innerPackConfig: '',
  palletConfig: '',
  // Dietary
  isVegan: false,
  isGlutenFree: false,
  isSugarFree: false,
  isKosher: false,
  isOrganic: false,
  // Sample Specs
  sampleMOQ: 0,
  sampleLeadTime: '',
  samplePrice: 0,
  translations: emptyTranslations()
} as any)

const localeTabLabelMap: Record<string, string> = {
  en: 'admin.products.locale_tab_en',
  zh: 'admin.products.locale_tab_zh',
  ko: 'admin.products.locale_tab_ko',
  ar: 'admin.products.locale_tab_ar',
  ja: 'admin.products.locale_tab_ja',
  th: 'admin.products.locale_tab_th',
  vi: 'admin.products.locale_tab_vi',
  id: 'admin.products.locale_tab_id',
  ms: 'admin.products.locale_tab_ms',
}

const localeTabLabel = (loc: string) => t(localeTabLabelMap[loc] || loc)

const parseCSV = (value: string) => value.split(',').map(v => v.trim()).filter(Boolean)

const buildTranslationsPayload = () => {
  const result: Record<string, Record<string, string>> = {}
  for (const loc of translationLocales) {
    const t = form.translations[loc]
    if (!t) continue
    const entry: Record<string, string> = {}
    if (t.name?.trim()) entry.name = t.name.trim()
    if (t.summary?.trim()) entry.summary = t.summary.trim()
    if (t.description?.trim()) entry.description = t.description.trim()
    if (Object.keys(entry).length > 0) result[loc] = entry
  }
  return Object.keys(result).length > 0 ? result : null
}

const aiTranslateAll = async () => {
  if (!editingId.value || aiTranslating.value) return
  aiTranslating.value = true
  try {
    const targetLocales = translationLocales.filter(l => l !== 'zh')
    const res = await api.post<any>(`/admin/products/${editingId.value}/ai-translate`, {
      productId: editingId.value,
      targetLocales,
    })
    if (res?.translations && typeof res.translations === 'object') {
      for (const loc of targetLocales) {
        const fields = res.translations[loc]
        if (fields && typeof fields === 'object') {
          if (fields.name) form.translations[loc].name = fields.name
          if (fields.summary) form.translations[loc].summary = fields.summary
          if (fields.description) form.translations[loc].description = fields.description
        }
      }
    }
  } catch (err: any) {
    formError.value = err?.message || t('errors.api.request_failed')
  } finally {
    aiTranslating.value = false
  }
}

const resetForm = () => {
  form.name = ''
  form.slug = ''
  form.summary = ''
  form.description = ''
  form.category = ''
  form.categorySlug = ''
  form.thumbnail = ''
  form.moq = 0
  form.basePrice = 0
  form.stockQuantity = 0
  form.leadTime = ''
  form.oemAvailable = false
  form.halalCertified = false
  form.featured = false
  form.status = 'active'
  form.imagesInput = ''
  form.flavorsInput = ''
  form.shapesInput = ''
  form.certificationsInput = ''
  form.ingredients = ''
  form.allergens = ''
  form.shelfLife = ''
  form.storage = ''
  form.netWeightPerPiece = 0
  form.netWeightPerPack = 0
  form.grossWeightPerCarton = 0
  form.piecesPerPack = 0
  form.packsPerCarton = 0
  form.productLengthMM = 0
  form.productWidthMM = 0
  form.productHeightMM = 0
  form.energyKj = 0
  form.energyKcal = 0
  form.totalFatG = 0
  form.saturatedFatG = 0
  form.carbohydratesG = 0
  form.sugarsG = 0
  form.proteinG = 0
  form.saltG = 0
  form.fiberG = 0
  form.additivesInput = ''
  form.sweetenerType = ''
  form.cocoaSolidsPct = 0
  form.milkSolidsPct = 0
  form.gmoStatus = ''
  form.mayContainInput = ''
  form.waterActivity = 0
  form.gtin = ''
  form.hsCode = ''
  form.primaryPackaging = ''
  form.innerPackConfig = ''
  form.palletConfig = ''
  form.isVegan = false
  form.isGlutenFree = false
  form.isSugarFree = false
  form.isKosher = false
  form.isOrganic = false
  form.sampleMOQ = 0
  form.sampleLeadTime = ''
  form.samplePrice = 0
  form.translations = emptyTranslations()
}

const fillFormFromProduct = (product: any) => {
  form.name = product.name || ''
  form.slug = product.slug || ''
  form.summary = product.summary || ''
  form.description = product.description || ''
  form.category = product.category || ''
  form.categorySlug = product.categorySlug || ''
  form.thumbnail = product.thumbnail || ''
  form.moq = product.moq || 0
  form.basePrice = product.basePrice || 0
  form.stockQuantity = product.stockQuantity || 0
  form.leadTime = product.leadTime || ''
  form.oemAvailable = Boolean(product.oemAvailable)
  form.halalCertified = Boolean(product.halalCertified)
  form.featured = Boolean(product.featured)
  form.status = product.status || 'active'
  form.imagesInput = Array.isArray(product.images) ? product.images.join(', ') : ''
  form.flavorsInput = Array.isArray(product.flavors) ? product.flavors.join(', ') : ''
  form.shapesInput = Array.isArray(product.shapes) ? product.shapes.join(', ') : ''
  form.certificationsInput = Array.isArray(product.certifications) ? product.certifications.join(', ') : ''
  form.ingredients = product.ingredients || ''
  form.allergens = product.allergens || ''
  form.shelfLife = product.shelfLife || ''
  form.storage = product.storage || ''
  form.netWeightPerPiece = product.netWeightPerPiece || 0
  form.netWeightPerPack = product.netWeightPerPack || 0
  form.grossWeightPerCarton = product.grossWeightPerCarton || 0
  form.piecesPerPack = product.piecesPerPack || 0
  form.packsPerCarton = product.packsPerCarton || 0
  form.productLengthMM = product.productLengthMM || 0
  form.productWidthMM = product.productWidthMM || 0
  form.productHeightMM = product.productHeightMM || 0
  form.energyKj = product.energyKj || 0
  form.energyKcal = product.energyKcal || 0
  form.totalFatG = product.totalFatG || 0
  form.saturatedFatG = product.saturatedFatG || 0
  form.carbohydratesG = product.carbohydratesG || 0
  form.sugarsG = product.sugarsG || 0
  form.proteinG = product.proteinG || 0
  form.saltG = product.saltG || 0
  form.fiberG = product.fiberG || 0
  form.additivesInput = Array.isArray(product.additives) ? product.additives.join(', ') : ''
  form.sweetenerType = product.sweetenerType || ''
  form.cocoaSolidsPct = product.cocoaSolidsPct || 0
  form.milkSolidsPct = product.milkSolidsPct || 0
  form.gmoStatus = product.gmoStatus || ''
  form.mayContainInput = Array.isArray(product.mayContain) ? product.mayContain.join(', ') : ''
  form.waterActivity = product.waterActivity || 0
  form.gtin = product.gtin || ''
  form.hsCode = product.hsCode || ''
  form.primaryPackaging = product.primaryPackaging || ''
  form.innerPackConfig = product.innerPackConfig || ''
  form.palletConfig = product.palletConfig || ''
  form.isVegan = Boolean(product.isVegan)
  form.isGlutenFree = Boolean(product.isGlutenFree)
  form.isSugarFree = Boolean(product.isSugarFree)
  form.isKosher = Boolean(product.isKosher)
  form.isOrganic = Boolean(product.isOrganic)
  form.sampleMOQ = product.sampleMOQ || 0
  form.sampleLeadTime = product.sampleLeadTime || ''
  form.samplePrice = product.samplePrice || 0
  if (product.translations && typeof product.translations === 'object') {
    const src = product.translations
    form.translations = emptyTranslations()
    for (const loc of translationLocales) {
      if (src[loc] && typeof src[loc] === 'object') {
        form.translations[loc].name = src[loc].name || ''
        form.translations[loc].summary = src[loc].summary || ''
        form.translations[loc].description = src[loc].description || ''
      }
    }
  } else {
    form.translations = emptyTranslations()
  }
}

const fetchProducts = async () => {
  pending.value = true; error.value = ''
  try {
    const res = await api.get<any>(`/admin/products?page=${page.value}&limit=${pageSize}`)
    products.value = res.data || []; pagination.value = res.pagination
  } catch (err: any) { error.value = err?.message || t('errors.api.load_failed') }
  finally { pending.value = false }
}

const nextPage = () => { if (pagination.value && page.value < pagination.value.totalPages) page.value += 1 }
const prevPage = () => { if (page.value > 1) page.value -= 1 }

const openCreateModal = () => { editingId.value = ''; translationLocale.value = 'en'; resetForm(); formError.value = ''; actionMessage.value = ''; actionError.value = false; showModal.value = true }

const openEditModal = async (product: any) => {
  editingId.value = product.id; formError.value = ''; actionMessage.value = ''; actionError.value = false
  try { const detail = await api.get<any>(`/admin/products/${product.id}`); fillFormFromProduct(detail); showModal.value = true }
  catch { fillFormFromProduct(product); showModal.value = true }
}

const closeModal = () => { showModal.value = false; saving.value = false; formError.value = '' }

const buildPayload = () => ({
  name: form.name, slug: form.slug, summary: form.summary, description: form.description,
  category: form.category, categorySlug: form.categorySlug, thumbnail: form.thumbnail,
  moq: form.moq, basePrice: Math.max(0, Number(form.basePrice) || 0), stockQuantity: Math.max(0, Number(form.stockQuantity) || 0), leadTime: form.leadTime,
  oemAvailable: form.oemAvailable, halalCertified: form.halalCertified, featured: form.featured, status: form.status,
  images: parseCSV(form.imagesInput), flavors: parseCSV(form.flavorsInput), shapes: parseCSV(form.shapesInput),
  certifications: parseCSV(form.certificationsInput), ingredients: form.ingredients, allergens: form.allergens,
  shelfLife: form.shelfLife, storage: form.storage,
  netWeightPerPiece: form.netWeightPerPiece, netWeightPerPack: form.netWeightPerPack, grossWeightPerCarton: form.grossWeightPerCarton,
  piecesPerPack: form.piecesPerPack, packsPerCarton: form.packsPerCarton,
  productLengthMM: form.productLengthMM, productWidthMM: form.productWidthMM, productHeightMM: form.productHeightMM,
  energyKj: form.energyKj, energyKcal: form.energyKcal, totalFatG: form.totalFatG, saturatedFatG: form.saturatedFatG,
  carbohydratesG: form.carbohydratesG, sugarsG: form.sugarsG, proteinG: form.proteinG, saltG: form.saltG, fiberG: form.fiberG,
  additives: parseCSV(form.additivesInput), sweetenerType: form.sweetenerType, cocoaSolidsPct: form.cocoaSolidsPct,
  milkSolidsPct: form.milkSolidsPct, gmoStatus: form.gmoStatus, mayContain: parseCSV(form.mayContainInput),
  waterActivity: form.waterActivity, gtin: form.gtin, hsCode: form.hsCode,
  primaryPackaging: form.primaryPackaging, innerPackConfig: form.innerPackConfig, palletConfig: form.palletConfig,
  isVegan: form.isVegan, isGlutenFree: form.isGlutenFree, isSugarFree: form.isSugarFree, isKosher: form.isKosher, isOrganic: form.isOrganic,
  sampleMOQ: form.sampleMOQ, sampleLeadTime: form.sampleLeadTime, samplePrice: form.samplePrice,
  translations: buildTranslationsPayload()
})

const saveProduct = async () => {
  if (!form.name.trim()) { formError.value = t('admin.products.name_required'); return }
  saving.value = true; formError.value = ''; actionMessage.value = ''; actionError.value = false
  const payload = buildPayload()
  try {
    if (editingId.value) { await api.put(`/admin/products/${editingId.value}`, payload); actionMessage.value = t('admin.products.updated_success') }
    else { await api.post('/admin/products', payload); actionMessage.value = t('admin.products.created_success') }
    closeModal(); await fetchProducts()
  } catch (err: any) { formError.value = err?.message || t('errors.api.save_failed') }
  finally { saving.value = false }
}

const deleteProduct = async (id: string) => {
  if (!confirm(t('admin.products.confirm_delete'))) return
  actionMessage.value = ''; actionError.value = false
  try { await api.del(`/admin/products/${id}`); actionMessage.value = t('admin.products.deleted_success'); await fetchProducts() }
  catch (err: any) { actionError.value = true; actionMessage.value = err?.message || t('errors.api.delete_failed') }
}

watch(page, fetchProducts)
onMounted(fetchProducts)
</script>

<style scoped>
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--spacing-xl);
}

.page-title {
  font-size: var(--text-2xl);
  font-weight: 600;
  color: var(--color-primary);
}

.page-subtitle {
  margin-top: var(--spacing-xs);
  font-size: var(--text-sm);
  color: var(--color-text-light);
}

.table-container {
  overflow: hidden;
  border-radius: var(--radius-lg);
  background: white;
  box-shadow: var(--shadow-md);
  margin-top: var(--spacing-xl);
}

.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table th {
  padding: var(--spacing-md);
  text-align: left;
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--color-primary);
  background: var(--color-bg-alt);
  border-bottom: 1px solid var(--color-border);
}

.data-table td {
  padding: var(--spacing-md);
  font-size: var(--text-sm);
  color: var(--color-text);
  border-bottom: 1px solid var(--color-border);
}

.data-table tbody tr:hover {
  background: var(--color-bg);
}

.data-table tbody tr:last-child td {
  border-bottom: none;
}

.pagination {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: var(--spacing-lg);
  padding: var(--spacing-md) var(--spacing-lg);
  background: white;
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-sm);
}

.link {
  color: var(--color-highlight);
  font-weight: 500;
  transition: color var(--transition-fast);
}

.link:hover {
  color: var(--color-highlight-hover);
}

.link-danger {
  color: var(--color-error);
}

.link-danger:hover {
  color: var(--color-error);
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--spacing-lg);
  margin-top: var(--spacing-lg);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
}

.form-group.col-span-2 {
  grid-column: span 2;
}

.form-label {
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-text);
}

.form-input {
  width: 100%;
  padding: var(--spacing-sm) var(--spacing-md);
  font-size: var(--text-base);
  color: var(--color-text);
  background: var(--color-bg);
  border: 1.5px solid var(--color-border);
  border-radius: var(--radius-md);
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
}

.form-input:focus {
  outline: none;
  border-color: var(--color-highlight);
  box-shadow: 0 0 0 3px rgba(var(--color-highlight-rgb), 0.1);
}

.form-textarea {
  width: 100%;
  padding: var(--spacing-sm) var(--spacing-md);
  font-size: var(--text-base);
  color: var(--color-text);
  background: var(--color-bg);
  border: 1.5px solid var(--color-border);
  border-radius: var(--radius-md);
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
  resize: vertical;
}

.form-textarea:focus {
  outline: none;
  border-color: var(--color-highlight);
  box-shadow: 0 0 0 3px rgba(var(--color-highlight-rgb), 0.1);
}

.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 50;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-xl);
}

.modal-container {
  position: relative;
  width: 100%;
  max-width: 800px;
  max-height: 90vh;
  overflow-y: auto;
}

.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
}

.modal-content {
  position: relative;
  background: white;
  border-radius: var(--radius-xl);
  padding: var(--spacing-xl);
  box-shadow: var(--shadow-xl);
  animation: modalIn 0.3s ease forwards;
}

@keyframes modalIn {
  from {
    opacity: 0;
    transform: scale(0.95) translateY(10px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.modal-title {
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-primary);
}
</style>
