export interface User {
  id: string;
  email: string;
  first_name?: string;
  last_name?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  roles: string[];
}

export interface Product {
  id: string;
  name: string;
  slug: string;
  description?: string;
  sku: string;
  price: number;
  compare_price?: number;
  cost_price?: number;
  track_inventory: boolean;
  weight?: number;
  status: 'draft' | 'active' | 'archived';
  published_at?: string;
  created_at: string;
  updated_at: string;
  variants: ProductVariant[];
  category_ids: string[];
  image_urls: string[];
}

export interface ProductVariant {
  id: string;
  product_id: string;
  title: string;
  sku: string;
  price: number;
  compare_price?: number;
  cost_price?: number;
  inventory_quantity: number;
  weight?: number;
  option1?: string;
  option2?: string;
  option3?: string;
  created_at: string;
  updated_at: string;
}

export interface Category {
  id: string;
  name: string;
  slug: string;
  description?: string;
  image_url?: string;
  parent_id?: string;
  sort_order: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Media {
  id: string;
  filename: string;
  original_name: string;
  mime_type: string;
  size_bytes: number;
  width?: number;
  height?: number;
  uploaded_by: string;
  created_at: string;
  url: string;
}

export interface Banner {
  id: string;
  title: string;
  image_url: string;
  link_url?: string;
  is_active: boolean;
  sort_order: number;
  starts_at?: string;
  ends_at?: string;
  created_at: string;
  updated_at: string;
}

export interface Offer {
  id: string;
  title: string;
  description?: string;
  discount_type: 'percentage' | 'fixed';
  discount_value: number;
  applies_to: 'all' | 'products' | 'categories' | 'collections';
  applies_to_ids: string[];
  min_order_amount?: number;
  usage_limit?: number;
  usage_count: number;
  is_active: boolean;
  starts_at: string;
  ends_at: string;
  created_at: string;
  updated_at: string;
}

export interface Order {
  id: string;
  order_number: string;
  user_id?: number;
  customer_name: string;
  customer_email: string;
  customer_phone: string;
  status: 'pending' | 'pending_payment' | 'processing' | 'shipped' | 'delivered' | 'cancelled' | 'paid' | 'just_arrived';
  currency: string;
  total_price: number;
  subtotal: number;
  tax_amount: number;
  shipping_amount: number;
  discount_amount: number;
  payment_method?: string;
  payment_id?: string;
  razorpay_order_id?: string;
  razorpay_payment_id?: string;
  razorpay_signature?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
  line_items: OrderLineItem[] | null;
  // Shipping fields
  shipping_name?: string;
  shipping_email?: string;
  shipping_phone?: string;
  shipping_address_line1?: string;
  shipping_address_line2?: string;
  shipping_city?: string;
  shipping_state?: string;
  shipping_country?: string;
  shipping_pin_code?: string;
  // Billing fields
  billing_name?: string;
  billing_email?: string;
  billing_phone?: string;
  billing_address_line1?: string;
  billing_address_line2?: string;
  billing_city?: string;
  billing_state?: string;
  billing_country?: string;
  billing_pin_code?: string;
  // Tracking fields
  tracking_number?: string;
  tracking_provider?: string;
  estimated_delivery?: string;
}

export interface OrderLineItem {
  id: string;
  order_id: string;
  product_id: string;
  variant_id?: string;
  title: string;
  sku: string;
  quantity: number;
  price: number;
  total: number;
}

export interface CreateOrderLineItem {
  product_id: string;
  variant_id?: string;
  quantity: number;
  price: number;
}

export interface Customer {
  id: string;
  email: string;
  first_name?: string;
  last_name?: string;
  phone?: string;
  addresses: Address[];
  created_at: string;
  updated_at: string;
}

export interface Address {
  id?: string;
  customer_id?: string;
  type: 'shipping' | 'billing';
  first_name: string;
  last_name: string;
  company?: string;
  address1: string;
  address2?: string;
  city: string;
  province?: string;
  country: string;
  postal_code: string;
  phone?: string;
}

export interface FormAddress {
  street?: string;
  city?: string;
  state?: string;
  country?: string;
  postal_code?: string;
}

export interface InventoryItem {
  id: string;
  product_id: string;
  variant_id?: string;
  sku: string;
  quantity: number;
  reserved: number;
  available: number;
  location?: string;
  cost_price?: number;
  updated_at: string;
}

export interface Setting {
  key: string;
  value: string;
  description?: string;
  type: 'string' | 'number' | 'boolean' | 'json';
  updated_at: string;
}

export interface AuditLog {
  id: string;
  user_id?: string;
  user_email?: string;
  action: string;
  resource_type: string;
  resource_id?: string;
  details?: Record<string, any>;
  ip_address?: string;
  user_agent?: string;
  created_at: string;
}
