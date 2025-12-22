export const SHIPPING_STATUS = {
  JUST_ARRIVED: 'just_arrived',
  PROCESSING: 'processing',
  SHIPPED: 'shipped',
  DELIVERED: 'delivered',
  CANCELLED: 'cancelled',
} as const;

export const SHIPPING_STATUS_OPTIONS = [
  { value: SHIPPING_STATUS.JUST_ARRIVED, label: 'Just Arrived' },
  { value: SHIPPING_STATUS.PROCESSING, label: 'Processing' },
  { value: SHIPPING_STATUS.SHIPPED, label: 'Shipped' },
  { value: SHIPPING_STATUS.DELIVERED, label: 'Delivered' },
  { value: SHIPPING_STATUS.CANCELLED, label: 'Cancelled' },
] as const;

export type ShippingStatus = typeof SHIPPING_STATUS[keyof typeof SHIPPING_STATUS];
