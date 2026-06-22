export interface CartEvent {
  eventId: number;
  domainName: string;
  listingId: number;
  inventoryType: number;
  itcCode: string;
  itcInventory: string;
  area: string;
  requestPrice: number;
  createdAt: string;
}
