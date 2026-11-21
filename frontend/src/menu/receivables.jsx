const receivablesMenu = {
  id: 'group-receivables',
  title: 'Receivables & Settlement',
  type: 'group',
  children: [
    {
      id: 'receivables-agenda',
      title: 'Receivables Agenda',
      type: 'item',
      url: '/receivables',
      icon: 'IconCalendarEvent'
    },
    {
      id: 'acquirer-connections',
      title: 'Acquirer Connections',
      type: 'item',
      url: '/receivables/connections',
      icon: 'IconCreditCard'
    },
    {
      id: 'matcher-reconciliation',
      title: 'Matcher Reconciliation',
      type: 'item',
      url: '/receivables/reconciliation',
      icon: 'IconArrowsShuffle'
    },
    {
      id: 'receivables-accounts',
      title: 'Accounts & Clearing',
      type: 'item',
      url: '/receivables/accounts',
      icon: 'IconBuildingBank'
    }
  ]
};

export default receivablesMenu;
