const receivablesMenu = {
  id: 'group-receivables',
  title: 'Gestão de Recebíveis',
  type: 'group',
  children: [
    {
      id: 'receivables-agenda',
      title: 'Agenda de Vendas',
      type: 'item',
      url: '/receivables',
      icon: 'IconCalendarEvent'
    },
    {
      id: 'acquirer-connections',
      title: 'Maquininhas e POS',
      type: 'item',
      url: '/receivables/connections',
      icon: 'IconCreditCard'
    },
    {
      id: 'matcher-reconciliation',
      title: 'Conciliação Bancária',
      type: 'item',
      url: '/receivables/reconciliation',
      icon: 'IconArrowsShuffle'
    },
    {
      id: 'receivables-accounts',
      title: 'Carteiras de Recebimento',
      type: 'item',
      url: '/receivables/accounts',
      icon: 'IconBuildingBank'
    }
  ]
};

export default receivablesMenu;
