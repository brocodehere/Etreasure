import React from 'react';
import { Route, Routes, Navigate, useLocation } from 'react-router-dom';
import { AdminLayout } from './shell/AdminLayout';
import { DashboardPage } from './pages/Dashboard';
import { LoginPage } from './pages/Login';
import { getAccessToken } from './lib/auth';
import { ProductsListPage } from './pages/Products/List';
import { MediaLibraryPage } from './pages/MediaLibrary';
import { ProductEditPage } from './pages/Products/Edit';
import { CategoriesListPage } from './pages/Categories/List';
import { BannersListPage } from './pages/Banners/List';
import { BannerEditPage } from './pages/Banners/Edit';
import { OffersListPage } from './pages/Offers/List';
import { OfferEditPage } from './pages/Offers/Edit';
import { OrdersListPage } from './pages/Orders/List';
import { SimpleOrdersPage } from './pages/Orders/SimpleList';
import { OrderDetailPage } from './pages/Orders/Detail';
import { OrderEditPage } from './pages/Orders/Edit';
import { CustomersListPage } from './pages/Customers/List';
import { CustomerEditPage } from './pages/Customers/Edit';
import { InventoryListPage } from './pages/Inventory/List';
import { InventoryEditPage } from './pages/Inventory/Edit';
import { SettingsListPage } from './pages/Settings/List';
import { SettingEditPage } from './pages/Settings/Edit';
import { UsersListPage } from './pages/Users/List';
import { UserEditPage } from './pages/Users/Edit';
import { AuditLogListPage } from './pages/AuditLog/List';
import { PreviewPage } from './pages/Preview';
import ContentManagement from './pages/Content/ContentManagement';
import ContentEdit from './pages/Content/ContentEdit';
import { FAQPage } from './pages/Content/FAQ';
import AboutPageManagement from './pages/Content/AboutPage';

const RequireAuth: React.FC<{ children: React.ReactElement }> = ({ children }) => {
  const location = useLocation();
  const token = getAccessToken();

  // Check if token exists and is not empty
  if (!token || token.trim() === '') {
    return <Navigate to="/login" replace state={{ from: location }} />;
  }

  // Optional: You could add token validation here
  // For now, just check if token exists
  return children;
};

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route
        path="/"
        element={
          <RequireAuth>
            <AdminLayout />
          </RequireAuth>
        }
      >
        <Route index element={<DashboardPage />} />
        <Route path="products">
          <Route index element={<ProductsListPage />} />
          <Route path="new" element={<ProductEditPage />} />
          <Route path=":id" element={<ProductEditPage />} />
        </Route>
        <Route path="media" element={<MediaLibraryPage />} />
        <Route path="categories">
          <Route index element={<CategoriesListPage />} />
          <Route path="new" element={<CategoriesListPage />} />
          <Route path=":id" element={<CategoriesListPage />} />
        </Route>
        <Route path="banners">
          <Route index element={<BannersListPage />} />
          <Route path="new" element={<BannerEditPage />} />
          <Route path=":id" element={<BannerEditPage />} />
        </Route>
        <Route path="offers">
          <Route index element={<OffersListPage />} />
          <Route path="new" element={<OfferEditPage />} />
          <Route path=":id" element={<OfferEditPage />} />
        </Route>
        <Route path="orders">
          <Route index element={<OrdersListPage />} />
          <Route path="new" element={<OrderEditPage />} />
          <Route path=":id" element={<OrderDetailPage />} />
          <Route path=":id/edit" element={<OrderEditPage />} />
        </Route>
        <Route path="customers">
          <Route index element={<CustomersListPage />} />
          <Route path="new" element={<CustomerEditPage />} />
          <Route path=":id" element={<CustomerEditPage />} />
        </Route>
        <Route path="inventory">
          <Route index element={<InventoryListPage />} />
          <Route path="new" element={<InventoryEditPage />} />
          <Route path=":id" element={<InventoryEditPage />} />
        </Route>
        <Route path="settings">
          <Route index element={<SettingsListPage />} />
          <Route path="new" element={<SettingEditPage />} />
          <Route path=":id" element={<SettingEditPage />} />
        </Route>
        <Route path="users">
          <Route index element={<UsersListPage />} />
          <Route path="new" element={<UserEditPage />} />
          <Route path=":id" element={<UserEditPage />} />
        </Route>
        <Route path="audit-log" element={<AuditLogListPage />} />
        <Route path="preview" element={<PreviewPage />} />
        <Route path="content">
          <Route index element={<ContentManagement />} />
          <Route path="faqs" element={<FAQPage />} />
          <Route path="about" element={<AboutPageManagement />} />
          <Route path="edit/:id" element={<ContentEdit />} />
          <Route path="new" element={<ContentEdit />} />
        </Route>
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
