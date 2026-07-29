export namespace excel {
	
	export class LabelData {
	    index: number;
	    title: string;
	    upc: string;
	
	    static createFrom(source: any = {}) {
	        return new LabelData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.title = source["title"];
	        this.upc = source["upc"];
	    }
	}
	export class LabelInfo {
	    sheet_map: Record<number, string>;
	    selected_sheet: number;
	    selected_sheet_name: string;
	    header_row: number;
	    header_col: string;
	    upc_col: string;
	    header_row_values: string[];
	    labels: LabelData[];
	    save_location: string;
	
	    static createFrom(source: any = {}) {
	        return new LabelInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sheet_map = source["sheet_map"];
	        this.selected_sheet = source["selected_sheet"];
	        this.selected_sheet_name = source["selected_sheet_name"];
	        this.header_row = source["header_row"];
	        this.header_col = source["header_col"];
	        this.upc_col = source["upc_col"];
	        this.header_row_values = source["header_row_values"];
	        this.labels = this.convertValues(source["labels"], LabelData);
	        this.save_location = source["save_location"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace image {
	
	export class Point {
	    X: number;
	    Y: number;
	
	    static createFrom(source: any = {}) {
	        return new Point(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.X = source["X"];
	        this.Y = source["Y"];
	    }
	}
	export class Rectangle {
	    Min: Point;
	    Max: Point;
	
	    static createFrom(source: any = {}) {
	        return new Rectangle(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Min = this.convertValues(source["Min"], Point);
	        this.Max = this.convertValues(source["Max"], Point);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RGBA {
	    Pix: number[];
	    Stride: number;
	    Rect: Rectangle;
	
	    static createFrom(source: any = {}) {
	        return new RGBA(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Pix = source["Pix"];
	        this.Stride = source["Stride"];
	        this.Rect = this.convertValues(source["Rect"], Rectangle);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace main {
	
	export class Placement {
	    height: number;
	    width: number;
	    origin_x: number;
	    origin_y: number;
	
	    static createFrom(source: any = {}) {
	        return new Placement(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.height = source["height"];
	        this.width = source["width"];
	        this.origin_x = source["origin_x"];
	        this.origin_y = source["origin_y"];
	    }
	}
	export class Layout {
	    image_height: number;
	    image_width: number;
	    barcode_placement: Placement;
	    title_placement: Placement;
	    page_height: number;
	    page_width: number;
	    ppi: number;
	
	    static createFrom(source: any = {}) {
	        return new Layout(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.image_height = source["image_height"];
	        this.image_width = source["image_width"];
	        this.barcode_placement = this.convertValues(source["barcode_placement"], Placement);
	        this.title_placement = this.convertValues(source["title_placement"], Placement);
	        this.page_height = source["page_height"];
	        this.page_width = source["page_width"];
	        this.ppi = source["ppi"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

