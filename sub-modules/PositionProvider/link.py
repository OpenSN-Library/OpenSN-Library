import uuid
from threading import RLock

class LinkEndInfo():

	def __init__(self,instance_id:str, instance_type:str):
		self.instance_id = ""
		self.instance_type = ""
		self.node_index = 0

	def __init__(self,obj: dict):
		for k,v in obj.items():
			setattr(self,k,v)

class LinkConfig:

	def __init__(self,link_index:int,link_type:str,init_parameter:dict,end_infos:list[LinkEndInfo]):
		self.link_index:int = 0
		self.type:str = link_type
		self.link_id:str = str(uuid.uuid4().hex)[:8]
		self.address_infos:list[dict[str,str]] = []
		self.init_parameter:dict[str,int] = init_parameter
		self.init_end_infos:list[LinkEndInfo] = end_infos

	def __init__(self):
		pass

	def __init__(self,obj: dict):
		for k,v in obj.items():
			if k == "init_end_infos":
				self.init_end_infos = []
				for item in v:
					self.init_end_infos.append(LinkEndInfo(item))
			else:
				setattr(self,k,v)

class LinkBase:
	def __init__(self,
			  link_id:str,
			  instance_id:list[str],
			  parameter:dict[str,int],
			  node_index:int,
			  link_index:int,
			  link_type:int,
			  init_parameter:dict,
			  end_infos:dict,
			):
		self.enabled = False
		self.cross_machine = instance_id[0] != instance_id[1]
		self.config = LinkConfig(node_index,link_index,link_type,init_parameter,end_infos)
		self.parameter:dict[str,int] = parameter

	def __init__(self):
		pass

	def __init__(self,obj: dict):
		for k,v in obj.items():
			if k == "config":
				self.config = LinkConfig(v)
			else:
				setattr(self,k,v)

Links : dict[str,LinkBase] = {}
LinksLock = RLock()

class ISL(LinkBase):
    
	def __init__(self, base: LinkBase):
		for k,v in base.__dict__.items:
			setattr(self,k,v)
		self.is_inter_orbit = False

	def __init__(self,obj: dict):
		LinkBase.__init__(self,obj)

class GSL(LinkBase):
    
	def __init__(self, base: LinkBase):
		for k,v in base.__dict__.items:
			setattr(self,k,v)

	def __init__(self,obj: dict):
		LinkBase.__init__(self,obj)

