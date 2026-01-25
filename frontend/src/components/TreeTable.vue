<template>
    <div class="tree-table-container">
        <table class="table table-sm table-bordered table-hover w-100 mb-0">
            <thead class="table-light">
                <tr>
                    <!-- Checkbox Column -->
                    <th class="text-center align-middle sticky-col sticky-col-checkbox" style="width: 40px">
                        <input type="checkbox" :checked="isAllSelected" @change="toggleSelectAll" />
                    </th>

                    <!-- Data Columns -->
                    <th v-for="field in fields" :key="field.Code"
                        :style="{ width: field.width ? field.width + 'px' : 'auto' }"
                        :class="{ 'sticky-col sticky-col-data': field.code === frozenColumnCode }">
                        {{ field.Name }}
                    </th>

                    <!-- Actions Column -->
                    <th class="actions text-center" style="width: 150px">
                        {{ $t('Actions') }}
                    </th>
                </tr>
            </thead>
            <tbody>
                <template v-if="flatTreeData.length > 0">
                    <tr v-for="row in flatTreeData" :key="row.id" :class="{ 'row-hidden': !row.isVisible }"
                        :data-level="row.level">
                        <!-- Checkbox -->
                        <td class="text-center align-middle sticky-col sticky-col-checkbox">
                            <input type="checkbox" :checked="selectedIds.includes(row.id)"
                                @change="toggleSelect(row.id)" />
                        </td>

                        <!-- Data Cells -->
                        <td v-for="(field, index) in fields" :key="field.Code"
                            :class="{ 'sticky-col sticky-col-data': field.code === frozenColumnCode }">

                            <!-- First Column: Indentation & Toggle -->
                            <div v-if="index === 0" class="d-flex align-items-center">
                                <!-- Indentation with Guides -->
                                <div class="d-flex" :style="{ width: (row.level - 1) * 24 + 'px' }">
                                    <span v-for="n in (row.level - 1)" :key="n" class="tree-indent-guide"
                                        :style="{ left: (n - 1) * 24 + 'px' }"></span>
                                </div>

                                <!-- Toggle Button -->
                                <span class="tree-toggle me-1 d-flex align-items-center justify-content-center"
                                    @click.stop="toggleExpand(row)"
                                    :class="{ 'is-leaf': !row.hasChildren, 'is-expanded': row.isExpanded }"
                                    style="width: 24px;">
                                    <i v-if="row.hasChildren" class="bi"
                                        :class="row.isExpanded ? 'bi-caret-down-fill' : 'bi-caret-right-fill'"></i>
                                    <i v-else class="bi bi-dot opacity-25"></i>
                                </span>

                                <!-- Value -->
                                <span class="text-truncate">
                                    <!-- status字段使用颜色标识 -->
                                    <template v-if="field.code === 'status'">
                                        <StatusBadge :status="row[field.code]" />
                                    </template>
                                    <template v-else>
                                        {{ row[`${field.code}_display`] !== null && row[`${field.code}_display`] !==
                                            undefined ?
                                            row[`${field.code}_display`] : row[field.code] }}
                                    </template>
                                </span>
                            </div>

                            <!-- Other Columns -->
                            <div v-else>
                                <!-- status字段使用颜色标识 -->
                                <template v-if="field.code === 'status'">
                                    <StatusBadge :status="row[field.code]" />
                                </template>
                                <template v-else>
                                    {{ row[`${field.code}_display`] !== null && row[`${field.code}_display`] !==
                                        undefined ?
                                        row[`${field.code}_display`] : '' }}
                                </template>
                            </div>
                        </td>

                        <!-- Actions -->
                        <td class="actions text-center">
                            <slot name="actions" :item="row"></slot>
                        </td>
                    </tr>
                </template>
                <tr v-else>
                    <td :colspan="fields.length + 2" class="text-center py-5 text-muted">
                        {{ $t('No Data') }}
                    </td>
                </tr>
            </tbody>
        </table>
    </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue';
import StatusBadge from '@/components/StatusBadge.vue';

const props = defineProps({
    items: {
        type: Array,
        default: () => [],
    },
    fields: {
        type: Array,
        default: () => [],
    },
    tableCode: {
        type: String,
        required: true,
    },
    frozenColumnCode: { // 支持冻结列
        type: String,
        default: ''
    }
});

const emit = defineEmits(['update:selected']);

const expandedIds = ref(new Set());
const selectedIds = ref([]);


// --- Tree Logic ---

// 1. Convert flat items to nested tree structure
const buildTree = (items) => {
    const map = {};
    const tree = [];
    const roots = [];

    // Initialize map and children arrays
    items.forEach(item => {
        map[item.id] = { ...item, children: [] };
    });

    // Build relationships
    items.forEach(item => {
        const node = map[item.id];
        if (item.parent_id && map[item.parent_id]) {
            map[item.parent_id].children.push(node);
        } else {
            roots.push(node);
        }
    });

    return roots;
};

// 2. Flatten tree into list for table rendering, respecting expansion state
const flattenTree = (nodes, level = 1, result = []) => {
    nodes.forEach(node => {
        const isExpanded = expandedIds.value.has(node.id);
        const hasChildren = node.children && node.children.length > 0;

        // Create specific row object for view
        result.push({
            ...node,
            level,
            isExpanded,
            hasChildren,
            isVisible: true // Root nodes are always visible (or dependent on parent which is handled by caller logic if we processed full tree)
        });

        if (hasChildren && isExpanded) {
            flattenTree(node.children, level + 1, result);
        }
    });
    return result;
};

// Computed flat tree data
const flatTreeData = computed(() => {
    if (!props.items || props.items.length === 0) return [];
    const roots = buildTree(props.items);
    return flattenTree(roots);
});

// --- Actions ---

const toggleExpand = (row) => {
    if (!row.hasChildren) return;

    if (expandedIds.value.has(row.id)) {
        expandedIds.value.delete(row.id);
    } else {
        expandedIds.value.add(row.id);
    }
    // Trigger reactivity strictly if needed, usually Set is not reactive in Vue 2 but ref(Set) works in Vue 3 if we re-assign or use reactive.
    // Wait, ref(Set) internal mutation might not trigger. Let's create a new Set.
    expandedIds.value = new Set(expandedIds.value);
};

const toggleSelect = (id) => {
    const index = selectedIds.value.indexOf(id);
    if (index > -1) {
        selectedIds.value.splice(index, 1);
    } else {
        selectedIds.value.push(id);
    }
    emit('update:selected', selectedIds.value);
};

const toggleSelectAll = (e) => {
    if (e.target.checked) {
        selectedIds.value = props.items.map(i => i.id);
    } else {
        selectedIds.value = [];
    }
    emit('update:selected', selectedIds.value);
};

const isAllSelected = computed(() => {
    return props.items.length > 0 && selectedIds.value.length === props.items.length;
});

// Initialize actions
// Default expand roots? Or all? Let's default expand level 1.
onMounted(() => {
    // auto expand logic if needed
});

// Watch for items change to reset or maintain state?
watch(() => props.items, (newItems) => {
    // maybe verify expandedIds exist in newItems
    // reset selectedIds? or keep?
});

</script>

<style scoped>
.tree-toggle {
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 20px;
    height: 20px;
    border-radius: 4px;
    transition: background-color 0.2s;
}

.tree-toggle:hover:not(.is-leaf) {
    background-color: rgba(0, 0, 0, 0.05);
}

.is-leaf {
    cursor: default;
}

/* Maintain sticky columns style from Index.vue */
.sticky-col {
    position: -webkit-sticky;
    position: sticky;
    background-color: white;
    z-index: 1;
}

.sticky-col-checkbox {
    left: 0;
    z-index: 3 !important;
}

.sticky-col-data {
    left: 40px;
    /* Adjust based on checkbox column width */
    z-index: 2 !important;
}

/* Ensure header sticky works too if the container scrolls */
thead .sticky-col {
    z-index: 4 !important;
    /* Higher than body sticky cols */
}

thead .sticky-col-checkbox {
    z-index: 5 !important;
}

/* Row hover effect */
tr:hover td {
    background-color: var(--bs-table-hover-bg);
}

/* Indent Guide Lines */
.tree-indent-guide {
    position: absolute;
    top: 0;
    bottom: 0;
    border-left: 1px dashed rgba(0, 0, 0, 0.1);
    width: 24px;
}

.tree-toggle i {
    transition: transform 0.2s;
    font-size: 0.8rem;
    color: #6c757d;
}

.tree-toggle.is-expanded i.bi-caret-right-fill {
    transform: rotate(90deg);
}

.d-flex.align-items-center {
    position: relative;
    height: 100%;
}
</style>
