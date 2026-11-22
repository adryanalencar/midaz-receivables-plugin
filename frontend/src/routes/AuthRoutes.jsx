import { Navigate } from 'react-router-dom';

const AuthRoutes = {
  path: 'login',
  element: <Navigate to="/receivables" replace />
};

export default AuthRoutes;
