<template>
  <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
    <!-- 标识字段（不可翻译） -->
    <div>
      <label for="product-slug" class="admin-form-label">{{ t('admin.products.slug') }}</label>
      <input id="product-slug" v-model="form.slug" name="slug" autocomplete="off" class="admin-form-control" />
    </div>
    <div>
      <label for="product-categorySlug" class="admin-form-label">{{ t('admin.products.category_slug') }}</label>
      <input id="product-categorySlug" v-model="form.categorySlug" name="categorySlug" autocomplete="off" class="admin-form-control" />
    </div>

    <!-- 多语言内容（唯一入口） -->
    <div class="admin-form-section sm:col-span-2">
      <div class="flex items-center justify-between mb-2">
        <div>
          <h4 class="admin-form-section-title">{{ t('admin.products.translations_label') }}</h4>
          <p class="mt-1 text-xs text-gray-500">{{ t('admin.products.translations_hint') }}</p>
        </div>
        <div v-if="editingId" class="inline-flex items-center">
          <button
            type="button"
            :disabled="aiTranslating"
            class="admin-btn-ghost"
            @click="$emit('aiTranslate')"
          >
            <Icon name="heroicons:sparkles" class="h-3.5 w-3.5" />
            {{ aiTranslating ? t('admin.products.ai_translating') : t('admin.products.ai_translate') }}
          </button>
          <AiHelpHint topic="products_translate" size="sm" />
        </div>
      </div>
      <div class="flex gap-2 mb-3 flex-wrap">
        <button
          v-for="loc in translationLocales"
          :key="loc"
          type="button"
          :class="['admin-locale-tab', translationLocale === loc ? 'admin-locale-tab--active' : '']"
          @click="$emit('update:translationLocale', loc)"
        >
          {{ localeTabLabel(loc) }}
        </button>
      </div>
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <div>
          <label :for="`trans-name-${translationLocale}`" class="admin-form-label">{{ t('admin.products.name') }} ({{ translationLocale }})</label>
          <input
            :id="`trans-name-${translationLocale}`"
            v-model="form.translations[translationLocale].name"
            :name="`transName-${translationLocale}`"
            autocomplete="off"
            :required="translationLocale === SOURCE_LOCALE"
            :placeholder="t('admin.products.trans_name_placeholder')"
            class="admin-form-control"
          />
        </div>
        <div>
          <label :for="`trans-alias-${translationLocale}`" class="admin-form-label">{{ t('admin.products.alias') }} ({{ translationLocale }})</label>
          <input :id="`trans-alias-${translationLocale}`" v-model="form.translations[translationLocale].alias" autocomplete="off" class="admin-form-control" />
        </div>
        <div>
          <label :for="`trans-category-${translationLocale}`" class="admin-form-label">{{ t('admin.products.category') }} ({{ translationLocale }})</label>
          <input :id="`trans-category-${translationLocale}`" v-model="form.translations[translationLocale].category" autocomplete="off" class="admin-form-control" />
        </div>
        <div>
          <label :for="`trans-categoryAlias-${translationLocale}`" class="admin-form-label">{{ t('admin.products.category_alias') }} ({{ translationLocale }})</label>
          <input :id="`trans-categoryAlias-${translationLocale}`" v-model="form.translations[translationLocale].categoryAlias" autocomplete="off" class="admin-form-control" />
        </div>
        <div class="sm:col-span-2">
          <label :for="`trans-summary-${translationLocale}`" class="admin-form-label">{{ t('admin.products.summary') }} ({{ translationLocale }})</label>
          <textarea :id="`trans-summary-${translationLocale}`" v-model="form.translations[translationLocale].summary" rows="2" autocomplete="off" class="admin-form-control"></textarea>
        </div>
        <div class="sm:col-span-2">
          <label :for="`trans-desc-${translationLocale}`" class="admin-form-label">{{ t('admin.products.description_label') }} ({{ translationLocale }})</label>
          <textarea :id="`trans-desc-${translationLocale}`" v-model="form.translations[translationLocale].description" rows="3" autocomplete="off" class="admin-form-control"></textarea>
        </div>
        <div>
          <label :for="`trans-leadTime-${translationLocale}`" class="admin-form-label">{{ t('admin.products.lead_time') }} ({{ translationLocale }})</label>
          <input :id="`trans-leadTime-${translationLocale}`" v-model="form.translations[translationLocale].leadTime" autocomplete="off" class="admin-form-control" />
        </div>
        <div>
          <label :for="`trans-flavors-${translationLocale}`" class="admin-form-label">{{ t('admin.products.flavors') }} ({{ translationLocale }})</label>
          <input :id="`trans-flavors-${translationLocale}`" v-model="form.translations[translationLocale].flavorsInput" autocomplete="off" class="admin-form-control" />
        </div>
        <div>
          <label :for="`trans-shapes-${translationLocale}`" class="admin-form-label">{{ t('admin.products.shapes') }} ({{ translationLocale }})</label>
          <input :id="`trans-shapes-${translationLocale}`" v-model="form.translations[translationLocale].shapesInput" autocomplete="off" class="admin-form-control" />
        </div>
        <div>
          <label :for="`trans-ingredients-${translationLocale}`" class="admin-form-label">{{ t('admin.products.ingredients') }} ({{ translationLocale }})</label>
          <textarea :id="`trans-ingredients-${translationLocale}`" v-model="form.translations[translationLocale].ingredients" rows="2" autocomplete="off" class="admin-form-control"></textarea>
        </div>
        <div>
          <label :for="`trans-allergens-${translationLocale}`" class="admin-form-label">{{ t('admin.products.allergens') }} ({{ translationLocale }})</label>
          <textarea :id="`trans-allergens-${translationLocale}`" v-model="form.translations[translationLocale].allergens" rows="2" autocomplete="off" class="admin-form-control"></textarea>
        </div>
        <div>
          <label :for="`trans-storage-${translationLocale}`" class="admin-form-label">{{ t('admin.products.storage') }} ({{ translationLocale }})</label>
          <input :id="`trans-storage-${translationLocale}`" v-model="form.translations[translationLocale].storage" autocomplete="off" class="admin-form-control" />
        </div>
        <div>
          <label :for="`trans-shelfLife-${translationLocale}`" class="admin-form-label">{{ t('admin.products.shelf_life') }} ({{ translationLocale }})</label>
          <input :id="`trans-shelfLife-${translationLocale}`" v-model="form.translations[translationLocale].shelfLife" autocomplete="off" class="admin-form-control" />
        </div>
      </div>
    </div>

    <!-- 基础信息（不可翻译） -->
    <div>
      <label for="product-moq" class="admin-form-label">{{ t('admin.products.moq') }}</label>
      <input id="product-moq" v-model.number="form.moq" name="moq" type="number" min="1" autocomplete="off" class="admin-form-control" />
    </div>
    <div>
      <label for="product-basePrice" class="admin-form-label">{{ t('admin.products.base_price') }}</label>
      <input id="product-basePrice" v-model.number="form.basePrice" name="basePrice" type="number" min="0.01" step="0.01" required autocomplete="off" class="admin-form-control" />
    </div>
    <div>
      <label for="product-stockQuantity" class="admin-form-label">{{ t('admin.products.stock_quantity') }}</label>
      <input id="product-stockQuantity" v-model.number="form.stockQuantity" name="stockQuantity" type="number" min="0" autocomplete="off" class="admin-form-control" />
    </div>
    <div>
      <label for="product-status" class="admin-form-label">{{ t('admin.products.status') }}</label>
      <select id="product-status" v-model="form.status" name="status" autocomplete="off" class="admin-form-control">
        <option value="active">{{ enumLabel('product_status', 'active') }}</option>
        <option value="draft">{{ enumLabel('product_status', 'draft') }}</option>
        <option value="inactive">{{ enumLabel('product_status', 'inactive') }}</option>
      </select>
    </div>
    <div class="flex items-center gap-4 pt-7">
      <label class="inline-flex items-center">
        <input id="product-oemAvailable" v-model="form.oemAvailable" name="oemAvailable" type="checkbox" class="admin-form-checkbox" />
        <span class="admin-form-check-label">{{ t('admin.products.oem') }}</span>
      </label>
      <label class="inline-flex items-center">
        <input id="product-halalCertified" v-model="form.halalCertified" name="halalCertified" type="checkbox" class="admin-form-checkbox" />
        <span class="admin-form-check-label">{{ t('admin.products.halal') }}</span>
      </label>
      <label class="inline-flex items-center">
        <input id="product-featured" v-model="form.featured" name="featured" type="checkbox" class="admin-form-checkbox" />
        <span class="admin-form-check-label">{{ t('admin.products.featured') }}</span>
      </label>
    </div>

    <!-- Weight & Measurement -->
    <div class="admin-form-section">
      <h4 class="admin-form-section-title">{{ t('admin.products.section_weight') }}</h4>
    </div>
    <div>
      <label for="product-netWeightPerPiece" class="admin-form-label">{{ t('admin.products.field_net_weight_per_piece') }}</label>
      <input id="product-netWeightPerPiece" v-model.number="form.netWeightPerPiece" type="number" step="0.1" class="admin-form-control" />
    </div>
    <div>
      <label for="product-netWeightPerPack" class="admin-form-label">{{ t('admin.products.field_net_weight_per_pack') }}</label>
      <input id="product-netWeightPerPack" v-model.number="form.netWeightPerPack" type="number" step="0.1" class="admin-form-control" />
    </div>
    <div>
      <label for="product-grossWeightPerCarton" class="admin-form-label">{{ t('admin.products.field_gross_weight_per_carton') }}</label>
      <input id="product-grossWeightPerCarton" v-model.number="form.grossWeightPerCarton" type="number" step="0.1" class="admin-form-control" />
    </div>
    <div>
      <label for="product-piecesPerPack" class="admin-form-label">{{ t('admin.products.field_pieces_per_pack') }}</label>
      <input id="product-piecesPerPack" v-model.number="form.piecesPerPack" type="number" class="admin-form-control" />
    </div>
    <div>
      <label for="product-packsPerCarton" class="admin-form-label">{{ t('admin.products.field_packs_per_carton') }}</label>
      <input id="product-packsPerCarton" v-model.number="form.packsPerCarton" type="number" class="admin-form-control" />
    </div>

    <!-- Dimensions -->
    <div class="admin-form-section">
      <h4 class="admin-form-section-title">{{ t('admin.products.section_dimensions') }}</h4>
    </div>
    <div>
      <label for="product-productLengthMM" class="admin-form-label">{{ t('admin.products.field_length_mm') }}</label>
      <input id="product-productLengthMM" v-model.number="form.productLengthMM" type="number" step="0.1" class="admin-form-control" />
    </div>
    <div>
      <label for="product-productWidthMM" class="admin-form-label">{{ t('admin.products.field_width_mm') }}</label>
      <input id="product-productWidthMM" v-model.number="form.productWidthMM" type="number" step="0.1" class="admin-form-control" />
    </div>
    <div>
      <label for="product-productHeightMM" class="admin-form-label">{{ t('admin.products.field_height_mm') }}</label>
      <input id="product-productHeightMM" v-model.number="form.productHeightMM" type="number" step="0.1" class="admin-form-control" />
    </div>

    <!-- Dietary -->
    <div class="admin-form-section">
      <h4 class="admin-form-section-title">{{ t('admin.products.section_dietary') }}</h4>
    </div>
    <div class="sm:col-span-2 flex items-center gap-4">
      <label class="inline-flex items-center">
        <input v-model="form.isVegan" type="checkbox" class="admin-form-checkbox" />
        <span class="admin-form-check-label">{{ t('admin.products.field_vegan') }}</span>
      </label>
      <label class="inline-flex items-center">
        <input v-model="form.isGlutenFree" type="checkbox" class="admin-form-checkbox" />
        <span class="admin-form-check-label">{{ t('admin.products.field_gluten_free') }}</span>
      </label>
      <label class="inline-flex items-center">
        <input v-model="form.isSugarFree" type="checkbox" class="admin-form-checkbox" />
        <span class="admin-form-check-label">{{ t('admin.products.field_sugar_free') }}</span>
      </label>
      <label class="inline-flex items-center">
        <input v-model="form.isKosher" type="checkbox" class="admin-form-checkbox" />
        <span class="admin-form-check-label">{{ t('admin.products.field_kosher') }}</span>
      </label>
      <label class="inline-flex items-center">
        <input v-model="form.isOrganic" type="checkbox" class="admin-form-checkbox" />
        <span class="admin-form-check-label">{{ t('admin.products.field_organic') }}</span>
      </label>
    </div>

    <!-- 媒体与认证 -->
    <div class="sm:col-span-2">
      <label for="product-thumbnail" class="admin-form-label">{{ t('admin.products.thumbnail') }}</label>
      <div class="mt-1 flex items-center gap-2">
        <input id="product-thumbnail" v-model="form.thumbnail" name="thumbnail" autocomplete="off" class="admin-form-control admin-form-control--flat" />
        <button type="button" class="admin-btn-ghost" @click="$emit('triggerUpload', 'thumbnail')">
          {{ t('admin.products.upload_image') }}
        </button>
      </div>
    </div>
    <div class="sm:col-span-2">
      <label for="product-ogImage" class="admin-form-label">{{ t('admin.products.og_image') }}</label>
      <p class="admin-form-hint mb-1">{{ t('admin.products.og_image_hint') }}</p>
      <div class="mt-1 flex items-center gap-2">
        <input id="product-ogImage" v-model="form.ogImage" name="ogImage" autocomplete="off" class="admin-form-control admin-form-control--flat" />
        <button type="button" class="admin-btn-ghost" @click="$emit('triggerUpload', 'ogImage')">
          {{ t('admin.products.upload_image') }}
        </button>
      </div>
    </div>
    <div class="sm:col-span-2">
      <label for="product-images" class="admin-form-label">{{ t('admin.products.images') }}</label>
      <div class="mt-1 flex items-center gap-2">
        <input id="product-images" v-model="form.imagesInput" name="images" autocomplete="off" class="admin-form-control admin-form-control--flat" />
        <button type="button" class="admin-btn-ghost" @click="$emit('triggerUpload', 'images')">
          {{ t('admin.products.upload_images') }}
        </button>
        <div v-if="uploadingImage" class="mt-1 admin-form-hint">
          <button type="button" class="admin-btn-danger-ghost" @click="$emit('cancelUpload')">
            {{ t('admin.products.cancel') }}
          </button>
        </div>
      </div>
    </div>
    <div class="sm:col-span-2">
      <label for="product-certifications" class="admin-form-label">{{ t('admin.products.certifications') }}</label>
      <input id="product-certifications" v-model="form.certificationsInput" name="certifications" autocomplete="off" class="admin-form-control" />
    </div>

    <!-- Nutrition per 100g -->
    <div class="admin-form-section">
      <h4 class="admin-form-section-title">{{ t('admin.products.section_nutrition') }}</h4>
    </div>
    <div>
      <label for="product-energyKj" class="admin-form-label">{{ t('admin.products.field_energy_kj') }}</label>
      <input id="product-energyKj" v-model.number="form.energyKj" type="number" step="1" class="admin-form-control" />
    </div>
    <div>
      <label for="product-energyKcal" class="admin-form-label">{{ t('admin.products.field_energy_kcal') }}</label>
      <input id="product-energyKcal" v-model.number="form.energyKcal" type="number" step="1" class="admin-form-control" />
    </div>
    <div>
      <label for="product-totalFatG" class="admin-form-label">{{ t('admin.products.field_fat_g') }}</label>
      <input id="product-totalFatG" v-model.number="form.totalFatG" type="number" step="0.1" class="admin-form-control" />
    </div>
    <div>
      <label for="product-saturatedFatG" class="admin-form-label">{{ t('admin.products.field_saturated_fat_g') }}</label>
      <input id="product-saturatedFatG" v-model.number="form.saturatedFatG" type="number" step="0.1" class="admin-form-control" />
    </div>
    <div>
      <label for="product-carbohydratesG" class="admin-form-label">{{ t('admin.products.field_carbohydrates_g') }}</label>
      <input id="product-carbohydratesG" v-model.number="form.carbohydratesG" type="number" step="0.1" class="admin-form-control" />
    </div>
    <div>
      <label for="product-sugarsG" class="admin-form-label">{{ t('admin.products.field_sugars_g') }}</label>
      <input id="product-sugarsG" v-model.number="form.sugarsG" type="number" step="0.1" class="admin-form-control" />
    </div>
    <div>
      <label for="product-proteinG" class="admin-form-label">{{ t('admin.products.field_protein_g') }}</label>
      <input id="product-proteinG" v-model.number="form.proteinG" type="number" step="0.1" class="admin-form-control" />
    </div>
    <div>
      <label for="product-saltG" class="admin-form-label">{{ t('admin.products.field_salt_g') }}</label>
      <input id="product-saltG" v-model.number="form.saltG" type="number" step="0.01" class="admin-form-control" />
    </div>
    <div>
      <label for="product-fiberG" class="admin-form-label">{{ t('admin.products.field_fiber_g') }}</label>
      <input id="product-fiberG" v-model.number="form.fiberG" type="number" step="0.1" class="admin-form-control" />
    </div>

    <!-- Ingredient Compliance -->
    <div class="admin-form-section">
      <h4 class="admin-form-section-title">{{ t('admin.products.section_compliance') }}</h4>
    </div>
    <div>
      <label for="product-additives" class="admin-form-label">{{ t('admin.products.field_additives') }}</label>
      <input id="product-additives" v-model="form.additivesInput" autocomplete="off" class="admin-form-control" />
    </div>
    <div>
      <label for="product-sweetenerType" class="admin-form-label">{{ t('admin.products.field_sweetener_type') }}</label>
      <input id="product-sweetenerType" v-model="form.sweetenerType" autocomplete="off" class="admin-form-control" />
    </div>
    <div>
      <label for="product-cocoaSolidsPct" class="admin-form-label">{{ t('admin.products.field_cocoa_solids_pct') }}</label>
      <input id="product-cocoaSolidsPct" v-model.number="form.cocoaSolidsPct" type="number" step="0.1" class="admin-form-control" />
    </div>
    <div>
      <label for="product-milkSolidsPct" class="admin-form-label">{{ t('admin.products.field_milk_solids_pct') }}</label>
      <input id="product-milkSolidsPct" v-model.number="form.milkSolidsPct" type="number" step="0.1" class="admin-form-control" />
    </div>
    <div>
      <label for="product-gmoStatus" class="admin-form-label">{{ t('admin.products.field_gmo_status') }}</label>
      <select id="product-gmoStatus" v-model="form.gmoStatus" class="admin-form-control">
        <option value="">—</option>
        <option value="GMO">{{ t('admin.products.gmo_option_gmo') }}</option>
        <option value="Non-GMO">{{ t('admin.products.gmo_option_non_gmo') }}</option>
        <option value="GMO-Free Certified">{{ t('admin.products.gmo_option_gmo_free_certified') }}</option>
      </select>
    </div>
    <div>
      <label for="product-mayContain" class="admin-form-label">{{ t('admin.products.field_may_contain') }}</label>
      <input id="product-mayContain" v-model="form.mayContainInput" autocomplete="off" class="admin-form-control" />
    </div>
    <div>
      <label for="product-waterActivity" class="admin-form-label">{{ t('admin.products.field_water_activity') }}</label>
      <input id="product-waterActivity" v-model.number="form.waterActivity" type="number" step="0.01" class="admin-form-control" />
    </div>

    <!-- Trade & Packaging -->
    <div class="admin-form-section">
      <h4 class="admin-form-section-title">{{ t('admin.products.section_trade') }}</h4>
    </div>
    <div>
      <label for="product-gtin" class="admin-form-label">{{ t('admin.products.field_gtin') }}</label>
      <input id="product-gtin" v-model="form.gtin" autocomplete="off" class="admin-form-control" />
    </div>
    <div>
      <label for="product-hsCode" class="admin-form-label">{{ t('admin.products.field_hs_code') }}</label>
      <input id="product-hsCode" v-model="form.hsCode" autocomplete="off" class="admin-form-control" />
    </div>
    <div>
      <label for="product-primaryPackaging" class="admin-form-label">{{ t('admin.products.field_primary_packaging') }}</label>
      <input id="product-primaryPackaging" v-model="form.primaryPackaging" autocomplete="off" :placeholder="t('admin.products.packaging_placeholder')" class="admin-form-control" />
    </div>
    <div>
      <label for="product-innerPackConfig" class="admin-form-label">{{ t('admin.products.field_inner_pack_config') }}</label>
      <input id="product-innerPackConfig" v-model="form.innerPackConfig" autocomplete="off" :placeholder="t('admin.products.inner_pack_placeholder')" class="admin-form-control" />
    </div>
    <div>
      <label for="product-palletConfig" class="admin-form-label">{{ t('admin.products.field_pallet_config') }}</label>
      <input id="product-palletConfig" v-model="form.palletConfig" autocomplete="off" :placeholder="t('admin.products.pallet_config_placeholder')" class="admin-form-control" />
    </div>

    <!-- Sample Specs -->
    <div class="admin-form-section">
      <h4 class="admin-form-section-title">{{ t('admin.products.section_sample') }}</h4>
    </div>
    <div>
      <label for="product-sampleMOQ" class="admin-form-label">{{ t('admin.products.field_sample_moq') }}</label>
      <input id="product-sampleMOQ" v-model.number="form.sampleMOQ" type="number" class="admin-form-control" />
    </div>
    <div>
      <label for="product-sampleLeadTime" class="admin-form-label">{{ t('admin.products.field_sample_lead_time') }}</label>
      <input id="product-sampleLeadTime" v-model="form.sampleLeadTime" autocomplete="off" class="admin-form-control" />
      <p class="mt-1 text-xs text-gray-500">{{ t('admin.products.field_sample_lead_time_hint') }}</p>
    </div>
    <div>
      <label for="product-samplePrice" class="admin-form-label">{{ t('admin.products.field_sample_price') }}</label>
      <input id="product-samplePrice" v-model.number="form.samplePrice" type="number" step="0.01" class="admin-form-control" />
    </div>

    <!-- Market profile & cost stack (edit only) -->
    <div v-if="editingId" class="sm:col-span-2 mt-4 border-t pt-4">
      <h4 class="admin-form-section-title mb-3">{{ t('admin.products.section_market') }}</h4>
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <div>
          <label class="admin-form-label">{{ t('admin.products.market_code') }}</label>
          <input v-model="marketForm.marketCode" class="admin-form-control uppercase" />
        </div>
        <div>
          <label class="admin-form-label">{{ t('admin.products.label_template') }}</label>
          <input v-model="marketForm.labelTemplateId" class="admin-form-control" />
        </div>
        <div class="sm:col-span-2">
          <label class="admin-form-label">{{ t('admin.products.destination_countries') }}</label>
          <input v-model="marketForm.destinationCountries" :placeholder="t('admin.products.csv_hint')" class="admin-form-control" />
        </div>
        <div class="sm:col-span-2">
          <label class="admin-form-label">{{ t('admin.products.required_certs') }}</label>
          <input v-model="marketForm.requiredCertKeywords" :placeholder="t('admin.products.csv_hint')" class="admin-form-control" />
        </div>
        <div class="sm:col-span-2 flex gap-2">
          <button type="button" class="admin-btn-primary" :disabled="marketSaving" @click="saveMarketProfile">{{ t('admin.products.save_market_profile') }}</button>
          <button type="button" class="admin-btn-outline" @click="loadMarketData">{{ t('admin.products.reload_market') }}</button>
        </div>
      </div>
      <div class="mt-6 border-t pt-4">
        <h4 class="admin-form-section-title mb-3">{{ t('admin.products.section_cost_stack') }}</h4>
        <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
          <div>
            <label class="admin-form-label admin-form-label--xs">{{ t('admin.products.cost_material') }}</label>
            <input v-model.number="costForm.materialCost" type="number" step="0.01" class="admin-form-control" />
          </div>
          <div>
            <label class="admin-form-label admin-form-label--xs">{{ t('admin.products.cost_labor') }}</label>
            <input v-model.number="costForm.laborCost" type="number" step="0.01" class="admin-form-control" />
          </div>
          <div>
            <label class="admin-form-label admin-form-label--xs">{{ t('admin.products.cost_logistics') }}</label>
            <input v-model.number="costForm.logisticsCost" type="number" step="0.01" class="admin-form-control" />
          </div>
        </div>
        <button type="button" class="mt-3 admin-btn-primary" :disabled="marketSaving" @click="saveMarketCost">{{ t('admin.products.save_cost_stack') }}</button>
        <p v-if="marketMessage" class="mt-2 text-xs" :class="marketError ? 'admin-form-msg--error' : 'admin-form-msg--success'">{{ marketMessage }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  form: any
  editingId: string
  aiTranslating: boolean
  translationLocale: string
  uploadingImage: boolean
}>()

defineEmits<{
  triggerUpload: [field: 'thumbnail' | 'ogImage' | 'images']
  cancelUpload: []
  aiTranslate: []
  'update:translationLocale': [locale: string]
}>()

const api = useApi()
const { t } = useI18n()
const { enumLabel } = useDisplay()

const marketForm = reactive({
  marketCode: 'EU', labelTemplateId: '', destinationCountries: '', requiredCertKeywords: '', notes: '',
})
const costForm = reactive({ materialCost: 0, laborCost: 0, logisticsCost: 0, marketCode: 'EU' })
const marketSaving = ref(false)
const marketMessage = ref('')
const marketError = ref(false)

const parseList = (s: string) => s.split(',').map(v => v.trim()).filter(Boolean)

/** 加载市场画像与成本栈 */
const loadMarketData = async () => {
  if (!props.editingId) return
  marketMessage.value = ''
  try {
    const profiles = await api.get<any[]>(`/admin/products/${props.editingId}/market-profile`, { marketCode: marketForm.marketCode })
    const p = Array.isArray(profiles) ? profiles[0] : profiles
    if (p) {
      marketForm.labelTemplateId = p.labelTemplateId || ''
      marketForm.notes = p.notes || ''
      try {
        const dc = typeof p.destinationCountries === 'string' ? JSON.parse(p.destinationCountries) : p.destinationCountries
        marketForm.destinationCountries = Array.isArray(dc) ? dc.join(', ') : ''
      } catch { /* ignore */ }
    }
    const costs = await api.get<any[]>(`/admin/products/${props.editingId}/market-costs`)
    const c = Array.isArray(costs) ? costs.find((x: any) => x.marketCode === marketForm.marketCode) || costs[0] : null
    if (c) {
      costForm.materialCost = c.materialCost ?? c.material_cost ?? 0
      costForm.laborCost = c.laborCost ?? c.labor_cost ?? 0
      costForm.logisticsCost = c.logisticsCost ?? c.logistics_cost ?? 0
    }
  } catch {
    marketMessage.value = t('admin.products.market_load_failed')
    marketError.value = true
  }
}

const saveMarketProfile = async () => {
  if (!props.editingId) return
  marketSaving.value = true
  marketMessage.value = ''
  marketError.value = false
  try {
    await api.put(`/admin/products/${props.editingId}/market-profile`, {
      marketCode: marketForm.marketCode,
      labelTemplateId: marketForm.labelTemplateId,
      destinationCountries: parseList(marketForm.destinationCountries),
      requiredCertKeywords: parseList(marketForm.requiredCertKeywords),
      notes: marketForm.notes,
    })
    marketMessage.value = t('admin.products.market_saved')
  } catch (err: any) {
    marketMessage.value = err?.message || t('errors.api.save_failed')
    marketError.value = true
  } finally {
    marketSaving.value = false
  }
}

const saveMarketCost = async () => {
  if (!props.editingId) return
  marketSaving.value = true
  marketMessage.value = ''
  marketError.value = false
  try {
    await api.put(`/admin/products/${props.editingId}/market-costs`, {
      marketCode: marketForm.marketCode,
      materialCost: costForm.materialCost,
      laborCost: costForm.laborCost,
      logisticsCost: costForm.logisticsCost,
    })
    marketMessage.value = t('admin.products.cost_saved')
  } catch (err: any) {
    marketMessage.value = err?.message || t('errors.api.save_failed')
    marketError.value = true
  } finally {
    marketSaving.value = false
  }
}

watch(() => props.editingId, (id) => { if (id) loadMarketData() }, { immediate: true })

// ALL_LOCALES / SOURCE_LOCALE 由 Nuxt 自动导入
const translationLocales = ALL_LOCALES

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
