import { lazy } from 'react';
import { Navigate } from 'react-router-dom';

// @project
import Loadable from '@/components/Loadable';
import { AuthProvider } from '@/contexts/AuthContext';
import AdminLayout from '@/layouts/AdminLayout';
import AuthGuard from '@/utils/route-guard/AuthGuard';
import RoleGuard from '@/utils/route-guard/RoleGuard';

// Views
const ReceivablesDashboard = Loadable(lazy(() => import('@/views/receivables/DashboardView')));
const ConnectionsView = Loadable(lazy(() => import('@/views/receivables/ConnectionsView')));
const ReconciliationView = Loadable(lazy(() => import('@/views/receivables/ReconciliationView')));
const AccountsView = Loadable(lazy(() => import('@/views/receivables/AccountsView')));

const MainRoutes = {
  path: '/',
  element: (
    <AuthProvider>
      <AuthGuard>
        <RoleGuard>
          <AdminLayout />
        </RoleGuard>
      </AuthGuard>
    </AuthProvider>
  ),
  children: [
    { index: true, element: <Navigate to="/receivables" replace /> },
    { path: 'receivables', element: <ReceivablesDashboard /> },
    { path: 'receivables/connections', element: <ConnectionsView /> },
    { path: 'receivables/reconciliation', element: <ReconciliationView /> },
    { path: 'receivables/accounts', element: <AccountsView /> }
  ]
};

export default MainRoutes;
