import { gql } from '@apollo/client';

export const GET_DASHBOARD_STATS = gql`
  query GetDashboardStats {
    dashboardStats {
      totalProducts
      totalCategories
      totalBanners
      activeBanners
      totalOffers
      activeOffers
      recentOrders
      totalRevenue
      outOfStockCount
    }
  }
`;

export const GET_JUST_ARRIVED_ORDERS = gql`
  query GetJustArrivedOrders {
    justArrivedOrders {
      id
      order_number
      customer_name
      customer_email
      total_price
      currency
      status
      shipping_status
      created_at
      updated_at
      line_items
    }
  }
`;
