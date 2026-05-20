<template>
  <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
    <!-- Basic Info -->
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
        <button type="button" class="rounded-md bg-orange-50 px-3 py-2 text-xs font-medium text-orange-600 hover:bg-orange-500" @click="$emit('triggerUpload', 'thumbnail')">
          {{ t('admin.products.upload_image') }}
        </button>
      </div>
    </div>
    <div class="sm:col-span-2">
      <label for="product-images" class="block text-sm font-medium text-gray-700">{{ t('admin.products.images') }}</label>
      <div class="mt-1 flex items-center gap-2">
        <input id="product-images" v-model="form.imagesInput" name="images" autocomplete="off" class="block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
        <button type="button" class="rounded-md bg-orange-50 px-3 py-2 text-xs font-medium text-orange-600 hover:text-orange-500" @click="$emit('triggerUpload', 'images')">
          {{ t('admin.products.upload_images') }}
        </button>
        <div v-if="uploadingImage" class="mt-1 text-xs text-gray-500">
          <button type="button" class="rounded-md bg-red-50 px-3 py-2 text-xs font-medium text-red-600 hover:text-red-500" @click="$emit('cancelUpload')">
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
        <option value="GMO">{{ t('admin.products.gmo_option_gmo') }}</option>
        <option value="Non-GMO">{{ t('admin.products.gmo_option_non_gmo') }}</option>
        <option value="GMO-Free Certified">{{ t('admin.products.gmo_option_gmo_free_certified') }}</option>
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
      <input id="product-primaryPackaging" v-model="form.primaryPackaging" autocomplete="off" :placeholder="t('admin.products.packaging_placeholder')" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
    </div>
    <div>
      <label for="product-innerPackConfig" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_inner_pack_config') }}</label>
      <input id="product-innerPackConfig" v-model="form.innerPackConfig" autocomplete="off" :placeholder="t('admin.products.inner_pack_placeholder')" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
    </div>
    <div>
      <label for="product-palletConfig" class="block text-sm font-medium text-gray-700">{{ t('admin.products.field_pallet_config') }}</label>
      <input id="product-palletConfig" v-model="form.palletConfig" autocomplete="off" :placeholder="t('admin.products.pallet_config_placeholder')" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
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
          @click="$emit('aiTranslate')"
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
          @click="$emit('update:translationLocale', loc)"
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
  </div>
</template>

<script setup lang="ts">
defineProps<{
  form: any
  editingId: string
  aiTranslating: boolean
  translationLocale: string
  uploadingImage: boolean
}>()

defineEmits<{
  triggerUpload: [field: 'thumbnail' | 'images']
  cancelUpload: []
  aiTranslate: []
  'update:translationLocale': [locale: string]
}>()

const { t } = useI18n()
const { enumLabel } = useDisplay()

const translationLocales = ['en', 'zh', 'ko', 'ar', 'ja', 'th', 'vi', 'id', 'ms']

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
</script>
