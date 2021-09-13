export const AIM_MD5_STRING = "AOL Instant Messenger (SM)";
export const FLAGS_EMPTY = Buffer.from([0x00, 0x00, 0x00, 0x00]);

export const enum USER_STATUS_VARIOUS {
  /**
   * Status webaware flag
   */
  WEBAWARE = 0x0001,
  /**
   * Status show ip flag
   */
  SHOWIP = 0x0002,
  /**
   * User birthday flag
   */
  BIRTHDAY = 0x0008,
  /**
   * User active webfront flag
   */
  WEBFRONT = 0x0020,
  /**
   * Direct connection not supported
   */
  DCDISABLED = 0x0100,
  /**
   * Direct connection upon authorization
   */
  DCAUTH = 0x1000,
  /**
   * DC only with contact users
   */
  DCCONT = 0x2000,
}

export const enum USER_STATUS {
  /**
   * Status is online
   */
  ONLINE = 0x0000,
  /**
   * Status is away
   */
  AWAY = 0x0001,
  /**
   * Status is no not disturb (DND)
   */
  DND = 0x0002,
  /**
   * Status is not available (N/A)
   */
  NA = 0x0004,
  /**
   * Status is occupied (BISY)
   */
  OCCUPIED = 0x0010,
  /**
   * Status is free for chat
   */
  FREE4CHAT = 0x0020,
  /**
   * Status is invisible
   */
  INVISIBLE = 0x0100,
}
