export namespace main {
	
	export class Item {
	    Title: string;
	    Link: string;
	    Description: string;
	    PubDate: string;
	    ImageURL: string;
	
	    static createFrom(source: any = {}) {
	        return new Item(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Title = source["Title"];
	        this.Link = source["Link"];
	        this.Description = source["Description"];
	        this.PubDate = source["PubDate"];
	        this.ImageURL = source["ImageURL"];
	    }
	}

}

