import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

// 导出一个函数，以便在任何组件的 setup 中调用
export function useFormOptions() {
  const { t } = useI18n();

  // 使用 computed 来确保当语言切换时，标签文本也能动态更新
  const statusOptions = computed(() => [
    { label: t('Normal'), value: 'Normal' },
    { label: t('Frozen'), value: 'Frozen' },
    { label: t('Deleted'), value: 'Deleted' },
  ]);

  const dataScopeOptions = computed(() => [
    { label: t('All Data'), value: 'All' },
    { label: t('Subordinate Data'), value: 'Subordinate' },
    { label: t('My Data'), value: 'Self' },
  ]);

  const protocolOptions = computed(() => [
    { label: t('Http'), value: 'Http' },
    { label: t('Rest'), value: 'Rest' },
    { label: t('GraphQL'), value: 'GraphQL' },
    { label: t('GRPC'), value: 'GRPC' },
    { label: t('Soap'), value: 'Soap' },
  ]);

  const methodOptions = computed(() => [
    { label: t('GET'), value: 'GET' },
    { label: t('POST'), value: 'POST' },
    { label: t('PUT'), value: 'PUT' },
    { label: t('DELETE'), value: 'DELETE' },
  ]);

  const platformOptions = [
    { label: t('Builtin'), value: 'Builtin' },
    { label: t('Feishu'), value: 'Feishu' },
    { label: t('DingTalk'), value: 'DingTalk' },
    { label: t('WeChatWork'), value: 'WeChatWork' },
    { label: t('Custom'), value: 'Custom' },
  ];

  const tableContentOptions = [
    { label: t('List'), value: 'List' },
    { label: t('Tree'), value: 'Tree' },
    // { label: t('Extension'), value: 'Extension' },
    // { label: t('Relation'), value: 'Relation' },
  ];

  const tableRelationOptions = [
    { label: t('Entity'), value: 'Entity' },
    { label: t('Item'), value: 'Item' },
  ];

  const fieldTypeOptions = [
    { label: t('Text'), value: 'Text' },
    { label: t('Date'), value: 'Date' },
    { label: t('Number'), value: 'Number' },
  ];

  const yesNoOptions = [
    { label: t('Yes'), value: 'Yes' },
    { label: t('No'), value: 'No' },
  ];

  const fieldStyleOptions = [
    { label: t('Text'), value: 'Text' },
    { label: t('Textarea'), value: 'Textarea' },
    { label: t('Integer'), value: 'Integer' },
    { label: t('Number'), value: 'Number' },
    { label: t('Checkbox'), value: 'Checkbox' },
    { label: t('Single Select'), value: 'Select' },
    { label: t('Multiple Select'), value: 'MultipleSelect' },
    { label: t('Radio Group'), value: 'RadioGroup' },
    { label: t('Checkbox Group'), value: 'CheckboxGroup' },
    { label: t('Attachment'), value: 'Attachment' },
    { label: t('Date'), value: 'Date' },
    { label: t('Time'), value: 'Time' },
    { label: t('Sequence'), value: 'Sequence' },
  ];

  const languageOptions = [
    { label: t('English'), value: 'en-us' },
    { label: t('Simplified Chinese'), value: 'zh-cn' },
    { label: t('Traditional Chinese'), value: 'zh-tw' },
    { label: t('Japanese'), value: 'ja-jp' },
    { label: t('Korean'), value: 'ko-kr' },
  ];

  const sexOptions = [
    { label: t('Male'), value: 'Male' },
    { label: t('Female'), value: 'Female' },
    { label: t('Other'), value: 'Other' },
  ];

  const headerContentTypeOptions = [
    { label: 'plain', value: 'plain' },
    { label: 'application/json', value: 'application/json' },
    { label: 'application/x-www-form-urlencoded', value: 'application/x-www-form-urlencoded' },
    { label: 'application/xml', value: 'application/xml' },
  ];

  // 审批实例状态常量
  // const (
  // ApprovalStatusPending = "Pending"  // 审批中
  // ApprovalStatusApproved = "Approved" // 已通过
  // ApprovalStatusRejected = "Rejected" // 已拒绝
  // ApprovalStatusCanceled = "Canceled" // 已撤回
  // ApprovalStatusDeleted = "Deleted"  // 已删除
  // ApprovalStatusExpired = "Expired"  // 已过期
  // 审批筛选选项 - 基于用户视角
  const approvalStatusOptions = [
    { label: t('Pending for Me'), value: 'Pending' },           // 待我审批的
    { label: t('Created by Me'), value: 'Created' },            // 我提交的
    { label: t('Processed by Me'), value: 'ProcessedByMe' },    // 我审批过的
    { label: t('All'), value: 'All' },                          // 全部
  ];

  // 新增时间范围筛选
  const timeRangeOptions = [
    { label: t('All'), value: 'All' },
    { label: t('Today'), value: 'Today' },
    { label: t('LastWeek'), value: 'LastWeek' },
    { label: t('LastMonth'), value: 'LastMonth' },
  ];

  const operationOptions = [
    { label: t('Create'), value: 'Create' },
    { label: t('Update'), value: 'Update' },
    { label: t('Delete'), value: 'Delete' },
    { label: t('Freeze'), value: 'Freeze' },
    { label: t('Unfreeze'), value: 'Unfreeze' },
    { label: t('Lock'), value: 'Lock' },
    { label: t('Unlock'), value: 'Unlock' },
    { label: t('Extend'), value: 'Extend' },
    { label: t('Void'), value: 'Void' },
    { label: t('Cancel'), value: 'Cancel' },
    { label: t('Terminate'), value: 'Terminate' },
    { label: t('BatchCreate'), value: 'BatchCreate' },
    { label: t('BatchUpdate'), value: 'BatchUpdate' },
    { label: t('BatchDelete'), value: 'BatchDelete' },
    { label: t('BatchFreeze'), value: 'BatchFreeze' },
    { label: t('BatchUnfreeze'), value: 'BatchUnfreeze' },
    { label: t('BatchLock'), value: 'BatchLock' },
    { label: t('BatchUnlock'), value: 'BatchUnlock' },
    { label: t('BatchExtend'), value: 'BatchExtend' },
  ];

  return {
    statusOptions,
    dataScopeOptions,
    protocolOptions,
    methodOptions,
    platformOptions,
    tableContentOptions,
    tableRelationOptions,
    fieldTypeOptions,
    yesNoOptions,
    fieldStyleOptions,
    languageOptions,
    sexOptions,
    headerContentTypeOptions,
    approvalStatusOptions,
    timeRangeOptions,
    operationOptions,
  };
}
